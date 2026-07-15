"""Flask server for benchmark comparison."""

from flask import Flask, jsonify, request

app = Flask(__name__)


@app.route("/")
def home():
    return jsonify({"message": "hello from Flask!"})


@app.route("/api/data")
def api_data():
    return jsonify({"items": [1, 2, 3], "service": "Flask"})


@app.route("/api/echo", methods=["POST"])
def api_echo():
    data = request.get_json()
    return jsonify(data)


if __name__ == "__main__":
    import sys

    port = int(sys.argv[1]) if len(sys.argv) > 1 else 8084
    app.run(host="0.0.0.0", port=port, debug=False)
