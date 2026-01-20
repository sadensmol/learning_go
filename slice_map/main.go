package main

import "fmt"

type TypeA struct {
	Name string
	Age  int
}

func main() {

	type Person struct {
		Name string
	}

	people := []Person{
		{Name: "Alice"},
		{Name: "Bob"},
	}

	fmt.Printf("%p\n", &people[0]) // Different address
	fmt.Printf("%p\n", &people[0]) // Different address

	// fmt.Print("--------------------------- testing arrays ---------------------------\n")
	// testArray()

	// fmt.Print("--------------------------- testing slices ---------------------------\n")
	// testSlice()

	// fmt.Print("--------------------------- testing maps ---------------------------\n")
	// testMap()

}

// you see here that every time it returns new item!!!
func testArray() {
	array := [4]TypeA{
		{Name: "Alice", Age: 30},
		{Name: "Bob", Age: 25},
		{Name: "Charlie", Age: 35},
	}

	fmt.Print("getting by range ...\n")
	for i, item := range array {
		fmt.Printf("item %d ( %p ), Name: %s, Age: %d\n", i, &item, item.Name, item.Age)
	}
	fmt.Print("getting by range 2nd time ...\n")
	for i, item := range array {
		fmt.Printf("item %d ( %p ), Name: %s, Age: %d\n", i, &item, item.Name, item.Age)
	}

	// getting by index
	fmt.Print("getting by index ...\n")
	for i := 0; i < len(array); i++ {
		fmt.Printf("item %d ( %p ), Name: %s, Age: %d\n", i, &array[i], array[i].Name, array[i].Age)
	}

	// getting by index
	fmt.Print("getting by index 2nd time...\n")
	for i := 0; i < len(array); i++ {
		fmt.Printf("item %d ( %p ), Name: %s, Age: %d\n", i, &array[i], array[i].Name, array[i].Age)
	}

	newItem := TypeA{Name: "New", Age: 20}
	fmt.Printf("new item ( %p ), Name: %s, Age: %d\n", &newItem, newItem.Name, newItem.Age)
	array[3] = newItem

	for i, item := range array {
		fmt.Printf("item %d ( %p ), Name: %s, Age: %d\n", i, &item, item.Name, item.Age)
	}

	fmt.Print("getting by range ...\n")
	for i, item := range array {
		fmt.Printf("item %d ( %p ), Name: %s, Age: %d\n", i, &item, item.Name, item.Age)
	}

	fmt.Print("getting by index ...\n")
	for i := 0; i < len(array); i++ {
		fmt.Printf("item %d ( %p ), Name: %s, Age: %d\n", i, &array[i], array[i].Name, array[i].Age)
	}
	fmt.Print("getting by index 2nd time...\n")
	for i := 0; i < len(array); i++ {
		fmt.Printf("item %d ( %p ), Name: %s, Age: %d\n", i, &array[i], array[i].Name, array[i].Age)
	}

	fmt.Print("changing array value in place ...\n")
	array[0].Age = 666
	array[1].Name = "New Bob"

	fmt.Print("getting by index after change ...\n")
	for i := 0; i < len(array); i++ {
		fmt.Printf("item %d ( %p ), Name: %s, Age: %d\n", i, &array[i], array[i].Name, array[i].Age)
	}

}

// range uses intenral copy now that's why we have copy every time!
func testSlice() {
	// create slice with anough capacity to hold 4 elements
	slice := make([]TypeA, 0, 4)
	fmt.Printf("initial slice( %p )\n", &slice)

	slice = append(slice, TypeA{Name: "Alice", Age: 30},
		TypeA{Name: "Bob", Age: 25},
		TypeA{Name: "Charlie", Age: 35},
	)
	fmt.Printf("initial slice after adding initial elements( %p )\n", &slice)
	fmt.Print("getting by range ...\n")
	for i, item := range slice {
		fmt.Printf("item %d ( %p ), Name: %s, Age: %d\n", i, &item, item.Name, item.Age)
	}

	fmt.Print("getting by range 2nd time ...\n")
	for i, item := range slice {
		fmt.Printf("item %d ( %p ), Name: %s, Age: %d\n", i, &item, item.Name, item.Age)
	}

	fmt.Print("getting by index ...\n")
	for i := 0; i < len(slice); i++ {
		fmt.Printf("item %d ( %p ), Name: %s, Age: %d\n", i, &slice[i], slice[i].Name, slice[i].Age)
	}

	fmt.Print("getting by index 2nd time ...\n")
	for i := 0; i < len(slice); i++ {
		fmt.Printf("item %d ( %p ), Name: %s, Age: %d\n", i, &slice[i], slice[i].Name, slice[i].Age)
	}

	newItem := TypeA{Name: "New", Age: 20}
	fmt.Printf("new item ( %p ), Name: %s, Age: %d\n", &newItem, newItem.Name, newItem.Age)
	fmt.Printf("slice before adding  ( %p )\n", &slice)
	// since we pass a copy of the new element to append function it will create a copy of it and stores this copy!
	slice = append(slice, newItem)
	fmt.Printf("slice after adding  ( %p )\n", &slice)

	fmt.Print("getting by range ...\n")
	for i, item := range slice {
		fmt.Printf("item %d ( %p ), Name: %s, Age: %d\n", i, &item, item.Name, item.Age)
	}

	fmt.Print("getting by index ...\n")
	for i := 0; i < len(slice); i++ {
		fmt.Printf("item %d ( %p ), Name: %s, Age: %d\n", i, &slice[i], slice[i].Name, slice[i].Age)
	}

	fmt.Print("getting by index 2nd time ...\n")
	for i := 0; i < len(slice); i++ {
		fmt.Printf("item %d ( %p ), Name: %s, Age: %d\n", i, &slice[i], slice[i].Name, slice[i].Age)
	}
}

// range uses internal copy for maps as well, similar to slices!
func testMap() {
	// create map with initial capacity
	m := make(map[string]TypeA, 4)
	fmt.Printf("initial map( %p )\n", &m)

	// add initial elements
	m["alice"] = TypeA{Name: "Alice", Age: 30}
	m["bob"] = TypeA{Name: "Bob", Age: 25}
	m["charlie"] = TypeA{Name: "Charlie", Age: 35}

	fmt.Printf("initial map after adding initial elements( %p )\n", &m)
	fmt.Print("getting by range ...\n")
	for key, item := range m {
		fmt.Printf("key: %s, item ( %p ), Name: %s, Age: %d\n", key, &item, item.Name, item.Age)
	}

	fmt.Print("getting by range 2nd time ...\n")
	for key, item := range m {
		fmt.Printf("key: %s, item ( %p ), Name: %s, Age: %d\n", key, &item, item.Name, item.Age)
	}

	fmt.Print("getting by key ...\n")
	for _, key := range []string{"alice", "bob", "charlie"} {
		item := m[key]
		fmt.Printf("key: %s, item ( %p ), Name: %s, Age: %d\n", key, &item, item.Name, item.Age)
	}

	fmt.Print("getting by key 2nd time ...\n")
	for _, key := range []string{"alice", "bob", "charlie"} {
		item := m[key]
		fmt.Printf("key: %s, item ( %p ), Name: %s, Age: %d\n", key, &item, item.Name, item.Age)
	}

	newItem := TypeA{Name: "New", Age: 20}
	fmt.Printf("new item ( %p ), Name: %s, Age: %d\n", &newItem, newItem.Name, newItem.Age)
	fmt.Printf("map before adding  ( %p )\n", &m)
	// since we pass a copy of the new element to map assignment it will create a copy of it and stores this copy!
	m["new"] = newItem
	fmt.Printf("map after adding  ( %p )\n", &m)

	fmt.Print("getting by range ...\n")
	for key, item := range m {
		fmt.Printf("key: %s, item ( %p ), Name: %s, Age: %d\n", key, &item, item.Name, item.Age)
	}

	fmt.Print("getting by key ...\n")
	for _, key := range []string{"alice", "bob", "charlie", "new"} {
		item := m[key]
		fmt.Printf("key: %s, item ( %p ), Name: %s, Age: %d\n", key, &item, item.Name, item.Age)
	}

	fmt.Print("getting by key 2nd time ...\n")
	for _, key := range []string{"alice", "bob", "charlie", "new"} {
		item := m[key]
		fmt.Printf("key: %s, item ( %p ), Name: %s, Age: %d\n", key, &item, item.Name, item.Age)
	}
}
