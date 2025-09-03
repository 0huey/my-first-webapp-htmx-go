document.addEventListener("DOMContentLoaded", (event) => {
	document.body.addEventListener('htmx:beforeSwap', function(evt) {
		if (evt.detail.xhr.status === 409) {
			evt.detail.shouldSwap = true;
			evt.detail.isError = false;
		}
	});
})

/*
document.addEventListener('htmx:confirm', function(evt) {
	if (!evt.target.hasAttribute('confirm-with-sweet-alert')) return
	const question = evt.target.getAttribute('confirm-with-sweet-alert');
	evt.preventDefault();
	Swal.fire({
		title: "Are you sure?",
		text: question || "Are you sure you want to continue?",
		icon: "warning",
	}).then((confirmed) => {
		if (confirmed) {
			evt.detail.issueRequest(true); // true to skip the built-in window.confirm()
		}
	});
});
*/
