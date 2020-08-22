package deleteendpoints

import (
	"context"
	"fmt"

	key "CarbonizeGoRestAPI/apikeys"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

//DeleteUserData deletes the first and last name as well
//as the email in the database
func DeleteUserData(id string) string {
	opt := option.WithCredentialsFile(key.FirebaseKeyPath)
	ctx := context.Background()
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return "noDelete"
	}
	//connect to the client
	client, err := app.Firestore(ctx)
	if err != nil {
		return "noDelete"
	}
	fmt.Println(id)
	//need to delete user first name and last name and email from database
	_, errors := client.Collection("userPersonalData").Doc(id).Update(ctx, []firestore.Update{
		{
			Path:  "firstName",
			Value: firestore.Delete,
		},
		{
			Path:  "lastName",
			Value: firestore.Delete,
		},
		{
			Path:  "email",
			Value: firestore.Delete,
		},
	})
	if errors != nil {
		return "noDelete"
	}

	//now delete document from the database
	_, error2 := client.Collection("userPersonalData").Doc(id).Delete(ctx)
	if error2 != nil {
		return "noDelete"
	}

	//encode a result to send back to frontend if deletion occurred smoothly
	return "deleted"
}
