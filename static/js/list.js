const toSubmit = new Set();
function toggleItem(box) {
    const item = $.text(box[0]
        .firstElementChild.firstElementChild.firstElementChild.nextElementSibling.firstElementChild).trim();
    if (toSubmit.has(item)) {
        toSubmit.delete(item);
        $(box[0].firstElementChild).removeClass("deselected");
    } else {
        toSubmit.add(item);
        $(box[0].firstElementChild).addClass("deselected");
    }
}

function submit() {
    var xhr = new XMLHttpRequest();
    xhr.open("POST", 'http://' + window.location.host + '/submit/', true);
    xhr.setRequestHeader("Content-type", "application/x-www-form-urlencoded");
    xhr.onreadystatechange = function() {//Call a function when the state changes.
        if (xhr.readyState === XMLHttpRequest.DONE && xhr.status === 200) {
            if (xhr.responseText === "refresh") {
                window.location.reload();
            } else {
                alert("Submitted!");
            }
        }
        if (xhr.readyState === XMLHttpRequest.DONE && xhr.status === 500) {
            alert(xhr.responseText);
        }
    };
    xhr.send(JSON.stringify(Array.from(toSubmit)));
}

window.onload = function () {
    const resetLink = $("#reset");
    if (resetLink.data("creator") === Cookies.get("username")) {
        resetLink.toggle();
    }
    $("#welcome").text("Welcome, " + Cookies.get("username") + "!");
};

