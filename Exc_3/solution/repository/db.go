package repository

import (
	"ordersystem/model"
	"time"
)

type DatabaseHandler struct {
	// drinks represent all available drinks
	drinks []model.Drink
	// orders serves as order history
	orders []model.Order
}

// todo
func NewDatabaseHandler() *DatabaseHandler {
	// Init the drinks slice with some test data
	// drinks := ...
	drinks := []model.Drink{
		{ID: 1, Name: "Bonk: Atomic Punch", Description: "Now with Isotopes!", Price: 2.20},
		{ID: 2, Name: "Blackout Stout", Description: "A true test for the true drunkard.", Price: 3},
		{ID: 3, Name: "DuPont Special", Description: "The DuPont Approach to drinks.", Price: 6.7},
	}
	// Init orders slice with some test data
	orders := []model.Order{
		{DrinkID: 1, CreatedAt: time.Now(), Amount: 10},
		{DrinkID: 2, CreatedAt: time.Now(), Amount: 3},
		{DrinkID: 3, CreatedAt: time.Now(), Amount: 21},
	}

	return &DatabaseHandler{
		drinks: drinks,
		orders: orders,
	}
}

func (db *DatabaseHandler) GetDrinks() []model.Drink {
	return db.drinks
}

func (db *DatabaseHandler) GetOrders() []model.Order {
	return db.orders
}

// todo
func (db *DatabaseHandler) GetTotalledOrders() map[uint64]uint64 {
	// calculate total orders
	// key = DrinkID, value = Amount of orders
	// totalledOrders map[uint64]uint64
	totalledOrders := make(map[uint64]uint64)

	for _, order := range db.orders {
		totalledOrders[order.DrinkID] += uint64(order.Amount)
	}

	return totalledOrders
}

func (db *DatabaseHandler) AddOrder(order *model.Order) {
	// todo
	// add order to db.orders slice
	db.orders = append(db.orders, *order)
}
