import os
import jwt


secret_key = os.environ.get("JWT_SECRET")

def decode_token(encoded_token):
    try:
        return jwt.decode(encoded_token, secret_key, algorithms=["HS256"])
    except jwt.ExpiredSignatureError:
        print("Token has expired.")
    except jwt.DecodeError:
        print("Token decoding failed.")

