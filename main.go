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

	// Fetch keys from etcd
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	var resp *clientv3.GetResponse
	resp, err = cli.Get(ctx, "", clientv3.WithPrefix())
	cancel()
	if err != nil {
		log.Fatal(err)
	}

	// Create a new tview application
	app := tview.NewApplication()

	// Create a list to display the keys
	list := tview.NewList().ShowSecondaryText(false)
	for _, kv := range resp.Kvs {
		list.AddItem(string(kv.Key), "", 0, nil)
	}

	// Set up the application layout
	flex := tview.NewFlex().
		AddItem(list, 0, 1, true)

	// Set the root and run the application
	err = app.SetRoot(flex, true).Run()
	if err != nil {
		log.Fatal(err)
	}
}
