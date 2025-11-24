const nicknamesData = document.getElementById("nicknames-data")

function addNickname() {
	const nicknames = document.getElementById("nicknames")
	let li = document.createElement("li")
	let input = document.createElement("input")
	let span = document.createElement("span")
	input.type = "text"
	span.innerHTML = "\u00d7"
	span.className = "deleteBtn"

	li.appendChild(input);
	li.appendChild(span);
	nicknames.appendChild(li);
}


nicknamesData.addEventListener("click", function(event) {
	if (event.target.classList.contains("deleteBtn")) {
		event.target.parentElement.remove()
		return
	} else if (event.target.classList.contains("saveBtn")) {
		let nicknames = []
		for (let li of document.getElementById("nicknames").children) {
			nicknames.push(li.children[0].value)
		}
		fetch("/account?edit=nicknames", {
			method: "PATCH",
			headers: {
				"Content-Type": "application/json"
			},
			body: JSON.stringify(nicknames)
		}).then(response => {
			if (!response.ok) {
				console.error(response)
				alert("Failed to save nicknames")
				return
			}
			console.log(response)
			window.location.reload()
		})
		return
	}
});
