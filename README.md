# Carbonize REST API
## Project Summary

This program creates the Carbonize RESTful API and is a key component of the complete Carbonize application. Using their Android phone, users are able to take a picture of a food and provide the carbon footprint of the food as well as the caloric value. The Carbonize application consists the following components - 
1. Frontend mobile app developed using Java to create the interactive UI elements and XML to create the layouts with Google's Material Design components (it supports light and dark mode!).
2. REST API written entirely in Golang with the Gorilla Mux library used to handle endpoints. Firebase and firestore client libraries are used to help perform CRUD operations on the database.
3. Firebase - the NoSQL database that stores the user's weekly data, their personal information (first name, last name, and email), and storing the user's most recent image. 
4. Google's Vision API is used to recognize the food in an image through the user of the Vision API's client library.
5. Google's NLP API analyzes a string to extract the name of the food when finding a recipe of a food. For example, if a recipe calls for "minced garlic" or "chopped garlic", the NLP API will extract the word "garlic" from the phrase.

## Endpoints

1. Analyze Image
   - This PUT method takes in a user's id and a JSON object containing the download URL of the image. The JSON object gets decoded and its values get stored into the ImageBody struct. Since the download URL comes from a Google service (Firebase), I'm able to use Google's Vision API to detect the labels from the image using the NewImageFromURI and DetectLabels functions. After getting results from the API, I use a hash map to store the description name as a key and the score as the value. After storing all of these key-value pairs into the hash map, I filter the results by deleting certain keys if some substrings are seen in the key and reducing the score of values if substrings are noticed in the keys. By filtering the results, almost all of the description labels sent back to the user will be food names. After filtering the results, I search the hash map to see the name associated with the highest value, which gets returned in a JSON object as the best result, and every other result is stored in a JSON array as the next best results.
2. Analyze Food
   - This POST method takes in a user's uid and a JSON object containing the name of the dish along with the day of the week. First I perform a query on the database in the "dishToValue" collection to see if the dish name is a document in the collection. If the dish is a document in the collection, we send back a json object with the existence variable set to true, and two other variables in the JSON object for the calorie count and carbon footprint of the food. If the dish isn't a document in the collection, I send a JSON object back to the user telling the user that the food isn't a document in the collection, which means that they'll have to listen to the collection to see when the document is created.
   - In another goroutine I do the following: I web scrape recipes to determine common ingredients in a food (I use Google's NLP API to perform entity analysis to pick out the name of the food in the string so I don't get results like "minced garlic" but just "garlic"), the calorie count of the food, and the cooking time of the food. By using goroutines, wait groups, and mutexes, I'm able to web scrape over 20 websites and make around 200 API calls in under a minute. I asynchronously calculate the calorie count and the carbon footprint by spawning a goroutine for each calculation (so 2 goroutines total). After getting both values, I create 2 goroutines, one to create a document for the dish, containing the calorie count and carbon footprint, and the other goroutine updates the user's calorie intake and carbon footprint graph data for the day.
3. Check User Exists, Get Name, Create User, Get Graph Data, Reset Graph Data, Delete User
   - These functions perform standard CRUD operations on user's information in Firebase.

## Technology Decisions
The main decision to choose Golang for all of the backend has to do with its speed and simple way of doing concurrency. The web scrapers were originally written in Python, and would take around 4 minutes on my laptop and home WiFi to run. In comparison, writing a web scraper in Golang, without concurrency, ran in under two minutes and with concurrency, ran in under 30 seconds with the same setup. Since the REST API was deployed to AWS with Elastic Beanstalk, the running time of the web scraper is most likely shorter. Although converting types is more difficult in Golang, there are workarounds in the language.

## Program Dependencies
1. OS - MacOS v10.11.6
2. Golang - Go v1.14.1
3. Gorilla Mux v1.7.4 - routing and handling endpoints
4. Firebase/Firestore - connecting to the database and performing CRUD operations
5. Go-Colly - web scraping
6. Google Vision API - analyze food from the image
7. Google NLP API - entity analysis on phrases
8. AWS Elastic Beanstalk - provision server instance to run the REST API

## Running the REST API
To run the REST API, you can type the following command into the console, assuming you're in the right directory and CarbonizeGoRestAPI is in your GOPATH:
```bash
go run carbonAPI.go
```
Keep in mind that the endpoints won't work without proper credential files, and the post method to calculate the carbon footprint and calorie count of the food won't work since I haven't added those files to the repository.

## Future Improvements
The next steps in the program would be to perform better error handling instead of causing the program to shutdown if an error occurs when performing CRUD operations on the database. Another feature to add would be to handle an exercising feature so that the user can keep track of calories burned and view exercise videos.

## Video Demo
For the demo, I was running the server code on an old 2010 Macbook Pro which is why responses times were slower than normal.
[https://youtu.be/KDFO9I55SM4] (https://youtu.be/KDFO9I55SM4)
