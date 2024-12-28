package main

import (
	"context"
	"github.com/gdamore/tcell/v2"
	"log"
	"strings"
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

	// Create a TreeView to display the keys
	treeView := tview.NewTreeView().
		SetRoot(tview.NewTreeNode("Root")).
		SetCurrentNode(tview.NewTreeNode("Root"))

	// Create a TextView to display the value of the selected key
	valueView := tview.NewTextView().SetDynamicColors(true).SetWrap(true)

	// Create a Flex layout to hold the tree view and value view side-by-side
	flex := tview.NewFlex().
		AddItem(treeView, 0, 1, true).
		AddItem(valueView, 0, 2, false)

	// Function to add a key to the tree
	addKeyToTree := func(root *tview.TreeNode, key string, value []byte) {
		parts := strings.Split(key, "/")
		currentNode := root
		for _, part := range parts {
			found := false
			for _, child := range currentNode.GetChildren() {
				if child.GetText() == part {
					currentNode = child
					found = true
					break
				}
			}
			if !found {
				newNode := tview.NewTreeNode(part)
				currentNode.AddChild(newNode)
				currentNode = newNode
			}
		}
		currentNode.SetReference(value)
		currentNode.SetSelectedFunc(func() {
			valueView.SetText(string(value))
			app.SetFocus(valueView)
		})
	}

	// Function to fetch keys from etcd and update the tree view
	updateTreeView := func() {
		var resp *clientv3.GetResponse
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		resp, err = cli.Get(ctx, "", clientv3.WithPrefix())
		cancel()
		if err != nil {
			log.Println(err)
			return
		}

		rootNode := tview.NewTreeNode("Root")
		treeView.SetRoot(rootNode)

		for _, kv := range resp.Kvs {
			key := string(kv.Key)
			addKeyToTree(rootNode, key, kv.Value)
		}
	}

	updateTreeView()

	// Add key event handler to switch focus between TreeView and TextView
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			if app.GetFocus() == treeView {
				app.SetFocus(valueView)
			} else {
				updateTreeView()
				app.SetFocus(treeView)
			}
		}
		return event
	})

	// Set the root and run the application
	err = app.SetRoot(flex, true).Run()
	if err != nil {
		log.Fatal(err)
	}
}
