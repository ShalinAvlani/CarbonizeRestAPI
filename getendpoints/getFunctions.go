package getendpoints

import (
	key "CarbonizeGoRestAPI/apikeys"
	"context"
	"fmt"
	"log"

	firebase "firebase.google.com/go"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

//GetTheName gets the name of the user and returns
//a json object of their first name, last name, and email
func GetTheName(id string) map[string]interface{} {
	//initialize firebase client
	opt := option.WithCredentialsFile(key.FirebaseKeyPath)
	ctx := context.Background()
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Fatal(err)
	}
	//connect to the client
	client, err := app.Firestore(ctx)
	if err != nil {
		log.Fatal(err)
	}
	//get the user's info document
	nameData, error1 := client.Collection("userPersonalData").Doc(id).Get(ctx)
	if error1 != nil {
		log.Fatal(error1)
	}
	fmt.Println("getName function")
	//get data from nameData and store in json object to return to frontend
	result := nameData.Data()
	fmt.Println(result)
	fmt.Println()
	return result
}

//CheckTheUser checks if the user exists in the database
//returns "notExists" if there were no failures and the email
//isn't used in the database, else returns "exists" if there was
//a failure at any step or if the email does exist in the database
func CheckTheUser(email string) string {
	//initialize firebase client
	opt := option.WithCredentialsFile(key.FirebaseKeyPath)
	ctx := context.Background()
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return "exists"
	}
	//connect to the client
	client, err := app.Firestore(ctx)
	if err != nil {
		return "exists"
	}
	iter := client.Collection("userPersonalData").Where("email", "==", email).Documents(ctx)
	//append any data to a slice
	var docSlice []map[string]interface{}
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return "exists"
		}
		docSlice = append(docSlice, doc.Data())
	}
	fmt.Println(len(docSlice))
	//if there are any elements in the slice, then the user already exists in the db
	if len(docSlice) > 0 {
		return "exists"
	}
	return "notExists"
}

//GetGraphData gets the user data from the database and
//returns a json object of the graph data
func GetGraphData(id string) map[string]interface{} {
	opt := option.WithCredentialsFile(key.FirebaseKeyPath)
	ctx := context.Background()
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Fatal(err)
	}
	//connect to client
	client, err := app.Firestore(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(id)
	//get the document containing the graph data
	graphData, err := client.Collection("users").Doc(id).Get(ctx)
	if err != nil {
		fmt.Println(err)
	}
	//get the data from the document and return it
	result := graphData.Data()
	fmt.Println(result)
	fmt.Println("Reached here")
	fmt.Println()
	return result
}
