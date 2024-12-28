package main

import (
	"context"
	"log"
	"time"

	"github.com/rivo/tview"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func main() {
	// Create an etcd client
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()

	// Create a new tview application
	app := tview.NewApplication()

	// Create a list to display the keys
	list := tview.NewList().ShowSecondaryText(false)

	// Function to fetch keys from etcd and update the list
	updateList := func() {
		var resp *clientv3.GetResponse
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		resp, err = cli.Get(ctx, "", clientv3.WithPrefix())
		cancel()
		if err != nil {
			log.Println(err)
			return
		}

		list.Clear()
		for _, kv := range resp.Kvs {
			list.AddItem(string(kv.Key), "", 0, nil)
		}
	}

	// Set up a ticker to update the list every 2 seconds
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	go func() {
		for range ticker.C {
			app.QueueUpdateDraw(updateList)
		}
	}()

	// Set up the application layout
	flex := tview.NewFlex().
		AddItem(list, 0, 1, true)

	// Set the root and run the application
	if err = app.SetRoot(flex, true).Run(); err != nil {
		log.Fatal(err)
	}
}
