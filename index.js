"use strict";
const fs = require('fs');
var oriGroceryArr = fs.readFileSync('./lists/groceries', 'utf8').trim().split("\n");
var unwantedArr = fs.readFileSync('./lists/unwanted', 'utf8').trim().split("\n\n");
const TOTAL = 87.88;
const N = 4;
var balances = {
  "eitan": TOTAL/N,
  "eric": TOTAL/N,
  "goom": TOTAL/N,
  "gayvin": TOTAL/N
};
var groceries = {};
var groceryArr = [];
var temp = [];
for (let i = 0; i < oriGroceryArr.length; i++) {
  temp.push(oriGroceryArr[i]);
  if (temp.length == 3) {
    groceryArr.push(temp);
    temp = [];
  }
}
for (let i = 0; i < groceryArr.length; i++) {
  var item = groceryArr[i];
  if (item.length != 3) {
    console.log(item);
    throw "error parsing! item length not 5 lines";
  }
  var name = item[0].trim().toLowerCase();
  name = name.substring(0, name.indexOf("image") - 1).trim();
  groceries[name] = Number(item[2].replace(/[^0-9\.]+/g,""));
}
var unwanted = {};
for (let i = 0; i < unwantedArr.length; i++) {
  var unwantedList = unwantedArr[i];
  var unwantedListArr = unwantedList.split("\n"); // a person's list
  var person = unwantedListArr[0].toLowerCase();
  for (let i = 1; i < unwantedListArr.length; i++) {
    var item = unwantedListArr[i];
    if (!unwanted[item.toLowerCase()]) {
      unwanted[item.toLowerCase()] = [person];
    } else {
      unwanted[item.toLowerCase()].push(person);
    }
  }
}

for (var item in unwanted) {
  var people = unwanted[item];
  var P = groceries[item];
  if (!P) {
    console.log(item);
    throw "error calculating! item not found";
  }
  var M = people.length;
  for (var person in balances) {
    if (people.indexOf(person) != -1) {
      balances[person] -= P / N;
    } else {
      balances[person] += (M * P) / (N * (N - M));
    }
  }
}
console.log(balances);
