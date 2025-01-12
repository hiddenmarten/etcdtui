package main

import (
	"context"
	"github.com/gdamore/tcell/v2"
	"log"
	"os"
	"strings"
	"time"

	"github.com/rivo/tview"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	app       *tview.Application
	keysView  *tview.TreeView
	valueView *tview.TextView
	cli       *clientv3.Client
)

func main() {
	// Create an etcd client
	var err error
	cli, err = newEtcdClient()
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

	// Init the value view empty and update the keys view
	updateValueView("")
	updateKeysView()

	// Add key event handler to switch focus between TreeView and TextView
	app.SetInputCapture(switchOnTab)

	// Set the root and run the application
	err = app.SetRoot(flex, true).Run()
	if err != nil {
		log.Fatal(err)
	}
}

func newEtcdClient() (*clientv3.Client, error) {
	endpoints := strings.Split(os.Getenv("ETCDCTL_ENDPOINTS"), ",")
	if len(endpoints) == 0 || endpoints[0] == "" {
		endpoints = []string{"localhost:2379"} // Default if not set
	}

	return clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
}

func addKeyToTree(root *tview.TreeNode, key string) {
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

	// Set the selected function to update the value view
	currentNode.SetSelectedFunc(selectNode(currentNode, key))
}

func getEtcdKeys() ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	resp, err := cli.Get(ctx, "", clientv3.WithPrefix(), clientv3.WithKeysOnly())
	cancel()
	if err != nil {
		return nil, err
	}
	keys := make([]string, 0)
	for _, kv := range resp.Kvs {
		keys = append(keys, string(kv.Key))
	}
	return keys, nil
}

func getEtcdValue(key string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	resp, err := cli.Get(ctx, key)
	cancel()
	if err != nil {
		return "", err
	}
	if len(resp.Kvs) == 0 {
		return "", nil
	}
	return string(resp.Kvs[0].Value), nil
}

func updateKeysView() {
	// Get the keys from etcd
	keys, err := getEtcdKeys()
	if err != nil {
		log.Fatal(err)
	}

	// Update the tree view
	rootNode := tview.NewTreeNode("Root")
	keysView.SetRoot(rootNode).SetBorder(true).SetTitle("Keys")
	for _, key := range keys {
		addKeyToTree(rootNode, key)
	}

}

func updateValueView(text string) {
	valueView.SetText(text).SetBorder(true).SetTitle("Value")
}

func selectNode(node *tview.TreeNode, key string) func() {
	return func() {
		value, err := getEtcdValue(key)
		if err != nil {
			log.Fatal(err)
		}
		node.SetReference(value)
		updateValueView(value)
		app.SetFocus(valueView)
	}
}

func switchOnTab(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyTab:
		if app.GetFocus() == keysView {
			app.SetFocus(valueView)
		} else {
			updateKeysView()
			app.SetFocus(keysView)
		}
	}
	return event
}
