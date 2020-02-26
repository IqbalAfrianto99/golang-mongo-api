package main

import (
  "context"
  "fmt"
  "net/http"
  "time"
  "encoding/json"

  "github.com/gorilla/mux"
  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/bson/primitive"
  "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/mongo/options"
)

type Person struct {
    ID          primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
    Firstname   string  `json:"firstname,omitempty" bson:"firstname,omitempty"`
    Lastname    string `json:"lastname,omitempty" bson:"lastname,omitempty"`
}

var client *mongo.Client
var db string = "golang-sandbox"
var collectionName string = "people"

func CreatePersonEndpoint(response http.ResponseWriter, request *http.Request) {
  // Create user
  response.Header().Set("content-type", "application/json")
  var person Person
  _ = json.NewDecoder(request.Body).Decode(&person)

  collection := client.Database(db).Collection(collectionName)
  ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

  result, _ := collection.InsertOne(ctx, person)
  json.NewEncoder(response).Encode(result)
}
func GetPeopleEndpoint(response http.ResponseWriter, request *http.Request) {
  // List user
  response.Header().Set("content-type", "application/json")
  var people []Person
  collection := client.Database(db).Collection(collectionName)
  ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
  cursor, err := collection.Find(ctx, bson.M{})

  if err != nil {
    response.WriteHeader(http.StatusInternalServerError)
    response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
    return
  }

  defer cursor.Close(ctx)

  for cursor.Next(ctx) {
    var person Person
    cursor.Decode(&person)
    people = append(people, person)
  }

  if err := cursor.Err(); err != nil {
    response.WriteHeader(http.StatusInternalServerError)
    response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
    return
  }

  json.NewEncoder(response).Encode(people)
}

// GetPersonEndpoint
func GetPersonEndpoint(response http.ResponseWriter, request *http.Request) {
  response.Header().Set("content-type", "application/json")
  params := mux.Vars(request)
  id, _ := primitive.ObjectIDFromHex(params["id"])
  var person Person

  fmt.Println(params)

  collection := client.Database(db).Collection(collectionName)
  ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
  err := collection.FindOne(ctx, Person{ID: id}).Decode(&person)
  if err != nil {
    response.WriteHeader(http.StatusInternalServerError)
    response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
    return
  }

  json.NewEncoder(response).Encode(person)
}

func main() {
  fmt.Println("Starting rest api")
  ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
  clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
  client, _ = mongo.Connect(ctx, clientOptions)

  router := mux.NewRouter()
  router.HandleFunc("/person", CreatePersonEndpoint).Methods("POST")
  router.HandleFunc("/person", GetPeopleEndpoint).Methods("GET")
  router.HandleFunc("/person/{id}", GetPersonEndpoint).Methods("GET")
  http.ListenAndServe(":12345", router)
}
