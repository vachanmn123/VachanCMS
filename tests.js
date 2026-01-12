fetch("http://localhost:8080/vachanmn123/VCMS-Test/test-type", {
  method: "POST",
  headers: {
    "Content-Type": "application/json",
  },
  body: JSON.stringify({
    values: {
      testField: "Hello, World",
    },
  }),
});

// Test uploading a media file
const formData = new FormData();
formData.append(
  "file",
  new Blob(["This is test media content"], { type: "text/plain" }),
  "test.txt",
);

fetch("http://localhost:8080/vachanmn123/VCMS-Test/media", {
  method: "POST",
  body: formData,
})
  .then((response) => response.json())
  .then((data) => {
    console.log("Upload response:", data);
    const mediaId = data.id;

    // Test listing media
    fetch("http://localhost:8080/vachanmn123/VCMS-Test/media?page=1")
      .then((response) => response.json())
      .then((listData) => {
        console.log("List media response:", listData);

        // Test getting media by ID
        fetch(`http://localhost:8080/vachanmn123/VCMS-Test/media/${mediaId}`)
          .then((response) => {
            if (response.ok) {
              console.log("Get media by ID successful");
              return response.text();
            } else {
              throw new Error("Failed to get media");
            }
          })
          .then((content) => {
            console.log("Media content:", content);
          })
          .catch((error) => console.error("Error getting media:", error));
      })
      .catch((error) => console.error("Error listing media:", error));
  })
  .catch((error) => console.error("Error uploading media:", error));
