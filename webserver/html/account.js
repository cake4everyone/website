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
	}
});

nicknamesData.addEventListener("htmx:configRequest", function(event) {
	if (event.target.classList.contains("saveBtn")) {
		let nicknames = []
		for (let li of document.getElementById("nicknames").children) {
			nicknames.push(li.children[0].value)
		}
		event.detail.formData.append("body", JSON.stringify(nicknames))
		return
	}
});
