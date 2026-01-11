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
