package main

import "errors"

var balances = make(map[string]float64)

func calculate() error {
	n, total := float64(uploadInfo.N), uploadInfo.Total
	for person := range submitInfo.Ready.Vals {
		balances[person] = total / n
	}
	for item, people := range submitInfo.Unwanted {
		p, ok := uploadInfo.Groceries[item]
		if !ok {
			return errors.New("item in unwanted list not found")
		}
		m := float64(people.Length())
		count := uint8(0)
		for person := range balances {
			if people.Contains(person) {
				balances[person] -= p / n
				count++
			} else {
				balances[person] += (m * p) / (n * (n - m))
			}
		}
		if count == uploadInfo.N {
			return errors.New("apparently, no one wanted " + item)
		}
	}
	return nil
}
