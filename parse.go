package main

import (
	"errors"
	"golang.org/x/net/html"
	"strconv"
	"strings"
)

var list *html.Node
var names = []string{}
var costs = []float64{}

func traverseHtml(n *html.Node, fn func(node *html.Node) bool) {
	done := fn(n)
	if done {
		return
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		traverseHtml(c, fn)
	}
}

func parseCost(n *html.Node) (float64, error) {
	data := strings.TrimSpace(n.FirstChild.Data)
	if len(data) > 0 && data[:1] == "$" {
		cost, err := strconv.ParseFloat(data[1:], 64)
		if err == nil {
			return cost, nil
		}
	}
	return 0.0, errors.New("failed to get float from string")
}

func findList(n *html.Node) bool {
	if n.Type == html.ElementNode {
		for _, attribute := range n.Attr {
			if attribute.Key == "id" && attribute.Val == "checkout-total-price-field" {
				cost, err := parseCost(n)
				if cost != 0.0 && err == nil {
					uploadInfo.Total = cost
				}
			} else if attribute.Key == "class" && attribute.Val == "a-box-group" {
				list = n
				return true
			}
		}
	}
	return false
}

func addToGroceries(n *html.Node) bool {
	if n.Type == html.ElementNode {
		for _, attribute := range n.Attr {
			if attribute.Key == "class" && attribute.Val == "a-size-base-plus a-color-base" {
				name := strings.TrimSpace(n.FirstChild.Data)
				if name != "" {
					names = append(names, name)
					submitInfo.Unwanted[name] = StringSet{Vals: make(map[string]struct{})}
				}
			}
			if attribute.Key == "class" &&
				attribute.Val == "a-size-base-plus a-color-base a-text-bold a-nowrap" {
				cost, err := parseCost(n)
				if cost != 0.0 && err == nil {
					costs = append(costs, cost)
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
	return false
}

func FindList(n *html.Node) {
	traverseHtml(n, findList)
}

func AddToGroceries(n *html.Node) {
	traverseHtml(n, addToGroceries)
	for i, name := range names {
		uploadInfo.Groceries[name] = costs[i]
	}
}
