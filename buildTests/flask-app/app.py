from flask import Flask

app = Flask(__name__)

@app.get("/")
def home():
    return "Hello from Flask on WSL! 🚀"

if __name__ == "__main__":
    # 0.0.0.0 lets you hit it from Windows via http://localhost:5000
    app.run(host="0.0.0.0", port=5000, debug=True)
