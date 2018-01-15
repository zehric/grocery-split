package main

import (
	"errors"
	"golang.org/x/net/html"
	"strconv"
	"strings"
	"fmt"
)

var names = []string{}
var costs = []float64{}

func traverseHtml(n *html.Node, fn func(node *html.Node) (*html.Node, error)) (node *html.Node, err error) {
	node, err = fn(n)
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if node != nil || err != nil {
			return
		}
		node, err = traverseHtml(c, fn)
	}
	return
}

func parseCost(n *html.Node) (float64, error) {
	data := strings.TrimSpace(n.FirstChild.Data)
	if len(data) > 0 && data[:1] == "$" {
		if data[len(data)-1:] == "*" {
			data = data[:len(data)-1]
		}
		cost, err := strconv.ParseFloat(data[1:], 64)
		if err == nil {
			return cost, nil
		}
	}
	return 0.0, errors.New("failed to get float from string")
}

func findList(n *html.Node) (*html.Node, error) {
	if n.Type == html.ElementNode {
		for _, attribute := range n.Attr {
			if attribute.Key == "id" && attribute.Val == "checkout-total-price-field" {
				cost, err := parseCost(n)
				if err != nil {
					return nil, err
				}
				if cost != 0.0 {
					uploadInfo.Total = cost
				}
			} else if attribute.Key == "class" && attribute.Val == "a-box-group" {
				return n, nil
			}
		}
	}
	return nil, nil
}

func addToGroceries(n *html.Node) (*html.Node, error) {
	if n.Type == html.ElementNode {
		for _, attribute := range n.Attr {
			if attribute.Key == "class" {
				if attribute.Val == "a-size-base-plus a-color-base" {
					name := strings.TrimSpace(n.FirstChild.Data)
					if name != "" {
						names = append(names, name)
						submitInfo.Unwanted[name] = StringSet{Vals: make(map[string]struct{})}
					}
				}
				if attribute.Val == "a-size-base-plus a-color-base a-text-bold a-nowrap" ||
					attribute.Val == "a-size-base-plus a-text-bold" {
					cost, err := parseCost(n)
					if err != nil {
						return nil, err
					}
					if cost != 0.0 {
						costs = append(costs, cost)
					}
				}
			}
		}
		if n.Data == "a" && len(n.Attr) == 2 {
			n.Attr = n.Attr[:1]
			hrefAttr := html.Attribute{Key: "href", Val: "javascript:;"}
			clickAttr := html.Attribute{Key: "onclick", Val: "toggleItem($(this))"}
			n.Attr = append(n.Attr, clickAttr)
			n.Attr = append(n.Attr, hrefAttr)
		}
	}
	return nil, nil
}

func reset() {
	names = []string{}
	costs = []float64{}
	submitInfo.Unwanted = make(map[string]StringSet)
	uploadInfo.Total = 0.0
}

func FindList(n *html.Node) (*html.Node, error) {
	list, err := traverseHtml(n, findList)
	if err != nil {
		return nil, err
	}
	if list == nil {
		return nil, errors.New("could not find grocery list in given html")
	}
	return list, err
}

func AddToGroceries(n *html.Node) error {
	_, err := traverseHtml(n, addToGroceries)
	if err != nil {
		reset()
		return err
	}
	if len(names) != len(costs) {
		errstr := "mismatched grocery names and prices:\n" +
			"names: " + fmt.Sprint(names) + " (" + strconv.Itoa(len(names)) + ")\n" +
			"costs: " + fmt.Sprint(costs) + " (" + strconv.Itoa(len(costs)) + ")"
		reset()
		return errors.New(errstr)
	}
	for i, name := range names {
		uploadInfo.Groceries[name] = costs[i]
	}
	return nil
}
