package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Todo struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"` //binary json
	Completed bool               `json:"completed" `
	Body      string             `json:"body"`
}

var collection *mongo.Collection

func main() {
	fmt.Println("Hello, World!")

	if os.Getenv("ENV") != "production" {
		err := godotenv.Load(".env")
		if err != nil {
			log.Fatal("Error loading .env file", err)

		}
	}

	MONGODB_URI := os.Getenv("MONGODB_URI")
	clientOptions := options.Client().ApplyURI(MONGODB_URI)
	client, err := mongo.Connect(context.Background(), clientOptions)

	if err != nil {
		log.Fatal(err)
	}

	defer client.Disconnect(context.Background())

	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to MongoDB!")

	collection = client.Database("go_todo").Collection("todos")

	app := fiber.New()
	if os.Getenv("ENV") != "production" {
		app.Use(cors.New(cors.Config{
			AllowOrigins: "*", //http://localhost:5173
			AllowHeaders: "Origin, Content-Type, Accept",
		}))
	}

	app.Get("/api/todos", getTodos)
	app.Post("/api/todos", postTodos)
	app.Patch("/api/todos/:id", patchTodos)
	app.Delete("/api/todos/:id", deleteTodos)

	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	if os.Getenv("ENV") == "production" {
		app.Static("/", "./client/react/dist")
	}

	log.Fatal(app.Listen(":" + port))

}

func getTodos(c *fiber.Ctx) error {
	fmt.Println("getTodos")
	var todos []Todo
	cursor, err := collection.Find(context.Background(), bson.D{})
	if err != nil {
		return err

	}

	defer cursor.Close(context.Background()) // mcm mysql close conn

	for cursor.Next(context.Background()) {
		var todo Todo
		if err := cursor.Decode(&todo); err != nil {
			return err
		}
		todos = append(todos, todo)
	}

	return c.JSON(todos)
}

func postTodos(c *fiber.Ctx) error {
	fmt.Println("postTodos")
	todo := new(Todo) // similar like todo := *Todo
	if err := c.BodyParser(todo); err != nil {
		return err
	}

	if todo.Body == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Body is required"})
	}

	res, err := collection.InsertOne(context.Background(), todo)
	if err != nil {
		return err
	}
	todo.ID = res.InsertedID.(primitive.ObjectID)

	return c.Status(201).JSON(todo)

}

func patchTodos(c *fiber.Ctx) error {
	fmt.Println("patchTodos")
	id := c.Params("id")
	objId, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
	}
	filter := bson.M{"_id": objId}
	update := bson.M{"$set": bson.M{"completed": true}}
	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}

	return c.Status(200).JSON(fiber.Map{"success": true})
}

func deleteTodos(c *fiber.Ctx) error {
	fmt.Println("deleteTodos")
	id := c.Params("id")
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
	}
	filter := bson.M{"_id": objId}
	_, err = collection.DeleteOne(context.Background(), filter)

	if err != nil {
		return err
	}

	return c.Status(200).JSON(fiber.Map{"success": true})
}
