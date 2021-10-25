import time
from flask import Flask

# Import RedMiddleware from last9
from last9.wsgi.middleware import RedMiddleware

app = Flask(__name__)

# Only line that needs to be changed
app.wsgi_app = RedMiddleware(app)

@app.route("/name/<name>")
def hello_world(name):
    time.sleep(1)
    return "<p>Hello, %s!</p>" % name

@app.route("/static")
def hello_static():
    return "<p>Static</p>"

if __name__ == "__main__":
    app.run('127.0.0.1', '5000', debug=True)
