import sys
import json

request = json.load(sys.stdin)
response = ['hello', 'world']
json.dump(response, sys.stdout)