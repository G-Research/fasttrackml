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
      success: namespaceIndex,
      error: namespaceIndex
    });
  });
}

function createNamespace() {
  window.location = window.location.origin + '/admin/ns/new'
}

function editNamespace(id) {
  window.location = window.location.origin + '/admin/ns/' + id
}

function deleteNamespace(id) {
  // Perform a DELETE request using jQuery's $.ajax
  $.ajax({
    url: "/admin/ns/" + id,
    type: "DELETE",
    contentType: "application/json",
    success: namespaceIndex,
    error: namespaceIndex
  });
}

function namespaceIndex() {
  window.location = window.location.origin + '/admin/ns/'
}
