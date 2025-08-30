// server.js
const express = require("express");
const path = require("path");

const app = express();
const PORT = 3000;

// Serve static files (CSS, images, etc.) from "public"
app.use(express.static(path.join(__dirname, "public")));

// Home page route
app.get("/", (req, res) => {
  res.send(`
    <h1>Hello from Express 🚀</h1>
    <p>This is a simple Express web app.</p>
    <a href="/about">Go to About</a>
  `);
});

// About page route
app.get("/about", (req, res) => {
  res.send(`
    <h1>About Page</h1>
    <p>This page is served by Express.</p>
    <a href="/">Back Home</a>
  `);
});

// Start server
app.listen(PORT, () => {
  console.log(`✅ Server running at http://localhost:${PORT}`);
});
