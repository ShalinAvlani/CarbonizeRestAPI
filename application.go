package main

import (
	key "CarbonizeGoRestAPI/apikeys"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	delete "CarbonizeGoRestAPI/deleteendpoints"
	get "CarbonizeGoRestAPI/getendpoints"
	post "CarbonizeGoRestAPI/postendpoints"
	cs "CarbonizeGoRestAPI/postendpoints/caloriescraper"
	carboncalc "CarbonizeGoRestAPI/postendpoints/carboncalc"
	put "CarbonizeGoRestAPI/putendpoints"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"github.com/gorilla/mux"
	"google.golang.org/api/option"
)

//ImageBody is used to handle the json object sent by the frontend in the analyzeImage endpoint
type ImageBody struct {
	DownloadURL string `json:"downloadURL"`
}

//FoodBody is used to handle the json object sent by the frontend in the analyzeFood endpoint
type FoodBody struct {
	DishName string `json:"dishName"`
	Day      string `json:"day"`
}

//UserBody is used to handle the json object sent by the frontend in the createUser endpoint
type UserBody struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
}

//ResetBody is used to handle the json object sent by the frontend in the resetData endpoint
type ResetBody struct {
	DaySelection  []string `json:"days"`
	DataSelection []string `json:"data"`
}

func main() {

	r := mux.NewRouter()
	r.HandleFunc("/api/resetData/{id}", resetData).Methods("PUT")
	r.HandleFunc("/api/analyzeFood/{id}", analyzeFood).Methods("POST")
	r.HandleFunc("/api/checkUserExists/{email}", checkUser).Methods("GET")
	r.HandleFunc("/api/getName/{id}", getName).Methods("GET")
	r.HandleFunc("/api/createUser/{id}", createNewUser).Methods("POST")
	r.HandleFunc("/api/getGraphData/{id}", getGraphData).Methods("GET")
	r.HandleFunc("/api/analyzeImage/{id}", analyzeImage).Methods("PUT")
	r.HandleFunc("/api/deleteUser/{id}", deleteUser).Methods("DELETE")
	log.Fatal(http.ListenAndServe(":5000", r))
}

//function to set all of the user's graph data for the week to 0
func resetData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	var resetData ResetBody
	err := json.NewDecoder(r.Body).Decode(&resetData)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(params["id"])
	fmt.Println(resetData.DaySelection)
	fmt.Println(resetData.DataSelection)
	result := put.ResetGraphData(params["id"], resetData.DaySelection, resetData.DataSelection)
	json.NewEncoder(w).Encode(map[string]bool{
		"result": result,
	})
}

//check if the user already exists in the database
func checkUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	result := get.CheckTheUser(params["email"])
	json.NewEncoder(w).Encode(map[string]string{
		"result": result,
	})
}

//get the first name and last name of user
func getName(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	result := get.GetTheName(params["id"])
	json.NewEncoder(w).Encode(result)
}

//precondition: user already doesn't exist in database
//create new user in table after making a new user, for carbon, calorieIn, and calorieBurn
func createNewUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	var userInfo UserBody
	//decode the json object and store the values in userInfo
	err := json.NewDecoder(r.Body).Decode(&userInfo)
	if err != nil {
		fmt.Println("ERROR DECODING JSON OBJ FROM CREATE NEW USER")
	}
	result := post.CreateUser(params["id"], userInfo.FirstName, userInfo.LastName, userInfo.Email)
	json.NewEncoder(w).Encode(map[string]bool{
		"result": result,
	})
}

//get graph data from user
func getGraphData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	result := get.GetGraphData(params["id"])
	json.NewEncoder(w).Encode(result)
}

//analyze image to return carbon footprint
func analyzeImage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	var img ImageBody
	err := json.NewDecoder(r.Body).Decode(&img)
	if err != nil {
		fmt.Println("ERROR")
	}
	jsonMap := put.AnalyzeImage(params["id"], img.DownloadURL)
	fmt.Println(jsonMap)
	fmt.Println()
	json.NewEncoder(w).Encode(jsonMap)
}

//delete user's personal data
func deleteUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	result := delete.DeleteUserData(params["id"])
	json.NewEncoder(w).Encode(map[string]string{
		"result": result,
	})
}

//analyze a food to return its carbon footprint and calorie count when given the name of the dish as well as the day of the week
//in order to update the user's data for the day
func analyzeFood(w http.ResponseWriter, r *http.Request) {
	opt := option.WithCredentialsFile(key.FirebaseKeyPath)
	ctx := context.Background()
	//initialize to the app and connect to the client
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Fatal(err)
	}
	client, err := app.Firestore(ctx)
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	fmt.Println(params)
	//put json object into the food object
	var food FoodBody
	json.NewDecoder(r.Body).Decode(&food)
	food.DishName = strings.ToLower(food.DishName)
	result := post.FindFoodExistence(food.DishName)
	if result.Existence {
		//add calorie count and carbon footprint to the user's info
		_, err1 := client.Collection("users").Doc(params["id"]).Update(ctx, []firestore.Update{
			{Path: food.Day + "Carbon", Value: firestore.Increment(result.CarbonFootprint)}, //update carbon footprint value
			{Path: food.Day + "CalorieIn", Value: firestore.Increment(result.CalorieCount)}, //update calorie intake value
		})
		if err1 != nil {
			log.Fatal(err1)
		}
		//send json body with calorie count and carbon footprint of food, "result": true
		json.NewEncoder(w).Encode(map[string]interface{}{
			"result":  true,
			"calorie": result.CalorieCount,
			"carbon":  result.CarbonFootprint,
		})
	} else {
		go func() {
			//if the food is something in the ingredients
			//start asynchronous functions to find the carbon footprint and calorie count
			var calorieCount float64
			var carbonFootprint float64
			var wg sync.WaitGroup
			wg.Add(1)
			//get calorie count of food
			go func(food string) {
				calorieCount = cs.GetFoodCalorieCount(food)
				wg.Done()
			}(food.DishName)

			wg.Add(1)
			//calculate carbon footprint
			go func(food string) {
				//calculate carbon footprint and store in carbon footprint variable
				carbonFootprint = carboncalc.CalculateCarbonFootprint(food)
				wg.Done()
			}(food.DishName)
			wg.Wait()

			fmt.Println("Calorie count:", calorieCount)
			fmt.Println("Carbon footprint:", carbonFootprint)
			//after finding carbon footprint and calorie count, do two things asynchronously:
			//create entry in db with carbon footprint and calorie count

			go func() {
				_, err1 := client.Collection("dishToValue").Doc(food.DishName).Set(ctx, map[string]interface{}{
					"calorie": calorieCount,
					"carbon":  carbonFootprint,
				})
				if err1 != nil {
					log.Fatal(err1)
				}
			}()

			//update user's data of carbon footprint and value
			go func() {
				_, err2 := client.Collection("users").Doc(params["id"]).Update(ctx, []firestore.Update{
					{Path: food.Day + "CalorieIn", Value: firestore.Increment(calorieCount)},
					{Path: food.Day + "Carbon", Value: firestore.Increment(carbonFootprint)},
				})
				if err2 != nil {
					log.Fatal(err2)
				}
			}()
		}()

		//send json body that just says "result": false
		json.NewEncoder(w).Encode(map[string]bool{
			"result": false,
		})

	}
}
