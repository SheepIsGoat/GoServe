import io
import logging
from traceback import print_exc

from fastapi import FastAPI, UploadFile, File, HTTPException
from starlette.requests import Request

# from pdfreader import SimplePDFViewer
import pdfplumber


logging.basicConfig(level=logging.INFO)

app = FastAPI()

@app.post("/extract/text")
async def extract_text(request: Request, file: UploadFile = File(...)):
    star_log(request, file)
    if file.content_type != 'application/pdf':
        raise HTTPException(status_code=400, detail="Invalid file type. Please upload a PDF.")
    
    try:
        extracted_text = await extract_text_from_strings(file)
        logging.info(f"Extracted text: {extracted_text}")
        return {"extracted_text": extracted_text}
    except:
        logging.error(f"Failed to extract raw text: {print_exc()}")
        raise HTTPException(status_code=500, detail="Failed to parse file")

    raise HTTPException(status_code=500, detail="Internal server error")
    


async def extract_text_from_strings(file):
    file_bytes = await file.read()
    text = ''
    with pdfplumber.open(io.BytesIO(file_bytes)) as pdf:
        for page in pdf.pages:
            text += page.extract_text() + '\n'
    return text

def star_log(request: Request, file: UploadFile=None):
    logging.info(f"Request Headers: {request.headers}")

    if file is not None:
        logging.info(f"Uploaded File Content-Type: {file.content_type}")
        logging.info(f"Uploaded File Filename: {file.filename}")