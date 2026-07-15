import gohttp
import gojson


class Server:
    def __init__(self):
        self._id = gohttp.create_server()
        self._handler_counter = 0

    def route(self, path, method="GET"):
        def wrapper(fn):
            self._handler_counter += 1
            handler_id = self._handler_counter
            func_name = "_handler_%d" % handler_id
            globals()[func_name] = fn
            gohttp.add_route(self._id, method, path, handler_id)
            return fn

        return wrapper

    def run(self, addr=":8080"):
        gohttp.start_server(self._id, addr)
        print("Server running on", addr)


app = Server()


@app.route("/")
def home(body):
    return '{"message": "hello from pocketpy!"}'


@app.route("/api/data")
def api_data(body):
    return '{"items": [1, 2, 3], "service": "pocketpy"}'


@app.route("/api/echo", method="POST")
def api_echo(body):
    return gojson.dumps(body)


if __name__ == "__main__":
    app.run()
