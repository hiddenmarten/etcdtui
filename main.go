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

var (
	app       *tview.Application
	keysView  *tview.TreeView
	valueView *tview.TextView
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
	app = tview.NewApplication()

	// Create a TreeView to display the keysView
	keysView = tview.NewTreeView().
		SetRoot(tview.NewTreeNode("Root")).
		SetCurrentNode(tview.NewTreeNode("Root"))

	// Create a TextView to display the value of the selected key
	valueView = tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(true)

	// Create a Flex layout to hold the tree view and value view side-by-side
	flex := tview.NewFlex().
		AddItem(keysView, 0, 1, true).
		AddItem(valueView, 0, 2, false)

	// Init the value view empty
	updateValueView("")

	// Update the tree view
	var resp *clientv3.GetResponse
	resp, err = getEtcdData(cli)
	if err != nil {
		log.Fatal(err)
	}
	updateTreeView(resp)

	// Add key event handler to switch focus between TreeView and TextView
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			if app.GetFocus() == keysView {
				app.SetFocus(valueView)
			} else {
				resp, err = getEtcdData(cli)
				if err != nil {
					log.Fatal(err)
				}
				updateTreeView(resp)
				app.SetFocus(keysView)
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

func addKeyToTree(root *tview.TreeNode, key string, value []byte) {
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
		updateValueView(string(value))
		app.SetFocus(valueView)
	})
}

func getEtcdData(cli *clientv3.Client) (*clientv3.GetResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	resp, err := cli.Get(ctx, "", clientv3.WithPrefix())
	cancel()
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func updateTreeView(resp *clientv3.GetResponse) {

	rootNode := tview.NewTreeNode("Root")
	keysView.SetRoot(rootNode).SetBorder(true).SetTitle("Keys")

	for _, kv := range resp.Kvs {
		key := string(kv.Key)
		addKeyToTree(rootNode, key, kv.Value)
	}
}

func updateValueView(text string) {
	valueView.SetText(text).SetBorder(true).SetTitle("Value")
}
