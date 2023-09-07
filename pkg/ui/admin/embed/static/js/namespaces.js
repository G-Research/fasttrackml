function handleUpdateNamespace() {
  $("#updateForm").on("submit", function(event) {
    event.preventDefault(); // Prevent the default form submission

    // Get form data
    const formData = $(this).serializeArray();

    // Convert formData to a regular object
    const formDataObject = {};
    formData.forEach(function(entry) {
      formDataObject[entry.name] = entry.value;
    });

    // Perform a PUT request using jQuery's $.ajax
    $.ajax({
      url: "/admin/ns/" + formDataObject["id"],
      type: "PUT",
      contentType: "application/json",
      data: JSON.stringify(formDataObject), // Convert to JSON format
    }).done(handleResponse);
  });
}

function createNamespace() {
  redirectTo('/admin/ns/new');
}

function editNamespace(id) {
  redirectTo('/admin/ns/' + id);
}

function namespaceIndex() {
  redirectTo('/admin/ns/');
}

function redirectTo(path) {
  window.location = window.location.origin + path;
}

function deleteNamespace(id) {
  if (confirm("Are you sure?") != true ){
    return
  }
  // Perform a DELETE request using jQuery's $.ajax
  $.ajax({
    url: "/admin/ns/" + id,
    type: "DELETE",
    contentType: "application/json",
  }).done(handleResponse);
}

function handleResponse(data, jqxhr, status) {
  redirectTo('/admin/ns/'
	     + "?message=" + encodeURIComponent(data["message"])
	     + "&status=" + data["status"]);
}
