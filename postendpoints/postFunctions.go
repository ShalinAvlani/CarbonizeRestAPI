package postendpoints

import (
	key "CarbonizeGoRestAPI/apikeys"
	"context"
	"fmt"
	"log"
	"reflect"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

//DBResult is a struct to return data back to the rest api with all the info of the CalorieCount and CarbonFootprint
type DBResult struct {
	Existence       bool
	CalorieCount    float32
	CarbonFootprint float32
}

func getDaysList() []string {
	return []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}
}

func getDataChoices() []string {
	return []string{"Carbon", "CalorieIn", "CalorieBurn"}
}

//convert a number value (it gets passed as an interface) to a float32
func convertToFloat(v interface{}) float32 {
	if reflect.ValueOf(v).Type().ConvertibleTo(reflect.TypeOf(float32(1.3))) {
		return float32(reflect.ValueOf(v).Convert(reflect.TypeOf(float32(1.3))).Float())
	}
	return float32(0)
}

//CreateUser creates a user by initializing their day data
//to 0 and creating a document in userPersonalData to store
//the person's first name, last name, and email
func CreateUser(id string, firstName string, lastName string, email string) bool {
	//initialize firebase client
	opt := option.WithCredentialsFile(key.FirebaseKeyPath)
	ctx := context.Background()
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return false
	}
	//connect to client
	client, err := app.Firestore(ctx)
	if err != nil {
		return false
	}
	fmt.Println("create new user")
	fmt.Println(id)
	//initialize graph data for the user
	daysList := getDaysList()
	dataChoices := getDataChoices()
	for i := 0; i < len(daysList); i++ {
		for j := 0; j < len(dataChoices); j++ {
			_, errors := client.Collection("users").Doc(id).Set(ctx, map[string]float32{
				daysList[i] + dataChoices[j]: 0,
			}, firestore.MergeAll)
			if errors != nil {
				fmt.Println(errors)
				return false
			}
		}
	}
	//set the name data for the user as well
	_, error2 := client.Collection("userPersonalData").Doc(id).Set(ctx, map[string]string{
		"firstName": firstName,
		"lastName":  lastName,
		"email":     email,
	})
	if error2 != nil {
		return false
	}
	//if there are any errors, function returns false, but if all goes well, then it returns true
	return true
}

//FindFoodExistence checks if the food already exists in the db
//if it does, returns true, else returns false and data needs to be web scraped
func FindFoodExistence(dishName string) DBResult {
	fmt.Println(dishName)
	//initialize firebase client
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
	//get document of food in the "dishToValue" collection
	obj1, _ := client.Collection("dishToValue").Doc(dishName).Get(ctx)
	firstExist := obj1.Exists()
	result := DBResult{Existence: firstExist}
	if firstExist {
		//food does exist, so it has the calorie count as well as the carbon footprint of the food
		data1 := obj1.Data()
		result.CarbonFootprint = convertToFloat(data1["carbon"])
		result.CalorieCount = convertToFloat(data1["calorie"])
		return result
	}
	//check if the food is in the ingredients list
	obj2, _ := client.Collection("ingredientToValue").Doc(dishName).Get(ctx)
	secondExist := obj2.Exists()
	result.Existence = secondExist
	if secondExist {
		//food exists in the ingredient collection, so it has the calorie count and the carbon footprint of the food
		data2 := obj2.Data()
		result.CarbonFootprint = convertToFloat(data2["carbon"])
		result.CalorieCount = convertToFloat(data2["calories"])
		return result
	}
	return result
}
