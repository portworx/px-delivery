package lib

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/sessions"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

type Loyalist struct {
	FirstName   string
	LastName    string
	Email       string
	Password    string
	LoyaltyTier string
}

type PageData struct {
	PageTitle  string
	Message    string
	SiteAction string
	ActionPage string
}

var (
	key              = []byte("kefue-secret-198")
	store            = sessions.NewCookieStore(key)
	mongoHost string = os.Getenv("MONGO_HOST")
	mongoUser string = os.Getenv("MONGO_USER")
	mongoPass string = os.Getenv("MONGO_PASS")
	mongoTLS  string = os.Getenv("MONGO_TLS")
)

func comparePasswords(hashedPwd string, plainPwd []byte) bool {
	// Since we'll be getting the hashed password from the DB it
	// will be a string so we'll need to convert it to a byte slice
	byteHash := []byte(hashedPwd)
	err := bcrypt.CompareHashAndPassword(byteHash, plainPwd)
	if err != nil {
		log.Println(err)
		return false
	}

	return true
}

func redirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", 301)
}

func hashAndSalt(pwd []byte) string {
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.DefaultCost)
	if err != nil {
		log.Println(err)
	}
	return string(hash)
}

func registerUser(firstname string, lastname string, email string, password string) {
	client, err := getMongoClient(mongoHost, mongoUser, mongoPass, mongoTLS)
	collection := client.Database("porxbbq").Collection("registrations")

	password = hashAndSalt([]byte(password))
	//fmt.Println(password)
	entry := Loyalist{firstname, lastname, email, password, "silver"}

	insertResult, err := collection.InsertOne(context.TODO(), entry)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Inserted a Single Document: ", insertResult.InsertedID)

}

func opstatus(w http.ResponseWriter, r *http.Request, messageData PageData) {
	fmt.Println("method:", r.Method, "on URL:", r.URL)
	session, _ := store.Get(r, "cookie-name")
	t, _ := template.ParseFiles("./static/internal_opstatus.html")
	if r.Method == "POST" {
		if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
			t, _ = template.ParseFiles("./static/external_opstatus.html")
		}
		t.Execute(w, messageData)
	}
}

func loginopstatus(w http.ResponseWriter, r *http.Request, messageData PageData) {
	fmt.Println("method:", r.Method, "on URL:", r.URL)
	session, _ := store.Get(r, "cookie-name")
	t, _ := template.ParseFiles("./static/order.html")
	if r.Method == "POST" {
		if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
			t, _ = template.ParseFiles("./static/login.html")
		}
		t.Execute(w, messageData)
	}
}
func checkLoyalist(email string) (found bool) {
	client, err := getMongoClient(mongoHost, mongoUser, mongoPass, mongoTLS)
	collection := client.Database("porxbbq").Collection("registrations")
	emailFilter := bson.D{{"email", email}}

	var result Loyalist

	found = false
	err = collection.FindOne(context.TODO(), emailFilter).Decode(&result)
	if err == nil {
		found = true
	}

	fmt.Println("Account with email address " + email + " was found " + strconv.FormatBool(found))
	return found
}

func checkCredentials(email string, password string) (access bool) {
	client, err := getMongoClient(mongoHost, mongoUser, mongoPass, mongoTLS)
	collection := client.Database("porxbbq").Collection("registrations")

	//convert plain text password to hash and compare with existing hash values
	passwordBytes := []byte(password)
	loginFilter := bson.D{{"email", email}}

	var result Loyalist

	access = false
	err = collection.FindOne(context.TODO(), loginFilter).Decode(&result)
	if err == nil {
		access = true
	}

	access = comparePasswords(result.Password, passwordBytes)

	return (access)
}

// Page Handlers
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method, "on URL:", r.URL)
	if r.URL.Path != "/healthz" {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}

	fmt.Fprintf(w, "Ready")
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method, "on URL:", r.URL)
	session, _ := store.Get(r, "cookie-name")
	if r.Method == "GET" {
		t, _ := template.ParseFiles("./static/register.html")
		t.Execute(w, nil)
	} else {
		found := checkLoyalist(r.FormValue("email"))
		if found == true {
			statusData := PageData{
				PageTitle: "Registration Status",
				Message:   "This Hard Topper has already been registered in the loyalty program!",
			}
			//Display Operation Status Page to User
			opstatus(w, r, statusData)
		} else {
			r.ParseForm()
			statusData := PageData{
				PageTitle: "Registration Status",
				Message:   "The email address " + r.FormValue("email") + " has been registered into the Loyalty Program!",
			}

			//write to mongo
			firstname := r.FormValue("firstname")
			lastname := r.FormValue("lastname")
			email := r.FormValue("email")
			password := r.FormValue("password")

			registerUser(firstname, lastname, email, password)

			//Display Operation Status Page to User
			session.Values["authenticated"] = true
			session.Values["email"] = r.FormValue("email")
			session.Save(r, w)
			opstatus(w, r, statusData)
		}
	}
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method, "on URL:", r.URL)
	session, _ := store.Get(r, "cookie-name")

	if r.Method == "GET" {
		t, _ := template.ParseFiles("./static/login.html")
		t.Execute(w, nil)
	} else {
		found := checkLoyalist(r.FormValue("email"))
		if found == true {

			// Check Credentials
			access := checkCredentials(r.FormValue("email"), r.FormValue("password"))
			if access == true {
				fmt.Println("Access Granted")
				statusData := PageData{
					PageTitle: "Login Status",
					Message:   "You are now logged in to the Loyalty Program",
				}
				session.Values["authenticated"] = true
				session.Values["email"] = r.FormValue("email")
				session.Save(r, w)
				opstatus(w, r, statusData)
			} else {
				fmt.Println("Access Denied")
				statusData := PageData{
					PageTitle: "Login Status",
					Message:   "You could not be logged in with your account. Please try again!",
				}
				opstatus(w, r, statusData)
			}

		} else {
			fmt.Println("Account Not Found")
			statusData := PageData{
				PageTitle: "Login Status",
				Message:   "Your account could not be located. Please try another email address or register!",
			}
			opstatus(w, r, statusData)
		}
	}
}

// called if you attempt to order without being logged in first.
// The redirect in this handler is different from the the Order Handler and Login Handlers
func OrderLoginHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method, "on URL:", r.URL)
	session, _ := store.Get(r, "cookie-name")

	if r.Method == "GET" {
		t, _ := template.ParseFiles("./static/login.html")
		t.Execute(w, nil)
	} else {
		found := checkLoyalist(r.FormValue("email"))
		if found == true {

			// Check Credentials
			access := checkCredentials(r.FormValue("email"), r.FormValue("password"))
			if access == true {
				fmt.Println("Access Granted")
				statusData := PageData{
					PageTitle: "Login Status",
					Message:   "You are now logged in to the Loyalty Program",
				}
				session.Values["authenticated"] = true
				session.Values["email"] = r.FormValue("email")
				session.Save(r, w)
				loginopstatus(w, r, statusData)
			} else {
				fmt.Println("Access Denied")
				statusData := PageData{
					PageTitle: "Login Status",
					Message:   "You could not be logged in with your account. Please try again!",
				}
				loginopstatus(w, r, statusData)
			}

		} else {
			fmt.Println("Account Not Found")
			statusData := PageData{
				PageTitle: "Login Status",
				Message:   "Your account could not be located. Please try another email address or register!",
			}
			loginopstatus(w, r, statusData)
		}
	}
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method, "on URL:", r.URL)
	session, _ := store.Get(r, "cookie-name")

	// Revoke users authentication
	session.Values["authenticated"] = nil
	session.Values["email"] = nil
	session.Save(r, w)

	http.Redirect(w, r, "/", 302)
}

func LoyaltystatusHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method, "on URL:", r.URL)
	session, _ := store.Get(r, "cookie-name")

	// Check if user is authenticated
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		fmt.Println("User not Authenticated - Forbidden!")
		statusData := PageData{
			PageTitle:  "",
			Message:    "Forbidden!",
			SiteAction: "Login",
			ActionPage: "/login",
		}
		if r.Method == "GET" {
			t, _ := template.ParseFiles("./static/loyaltystatus.html")
			t.Execute(w, statusData)
		}
		return
	}

	client, err := getMongoClient(mongoHost, mongoUser, mongoPass, mongoTLS)
	collection := client.Database("porxbbq").Collection("registrations")

	loginFilter := bson.D{{"email", session.Values["email"]}}

	var result Loyalist

	err = collection.FindOne(context.TODO(), loginFilter).Decode(&result)
	if err != nil {
		log.Fatal(err)
	}

	statusData := PageData{
		PageTitle:  "Loyalty Status",
		Message:    "Thank you for being a loyalty customer. You're current status level is " + fmt.Sprintf("%v", result.LoyaltyTier) + "!",
		SiteAction: "Logout",
		ActionPage: "/logout",
	}
	if r.Method == "GET" {
		t, _ := template.ParseFiles("./static/loyaltystatus.html")
		t.Execute(w, statusData)
	}

}

// end of page handlers
