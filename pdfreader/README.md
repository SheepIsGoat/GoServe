
```
curl -X POST "http://0.0.0.0:8000/extract_text" \
     -H "accept: application/json" \
     -H "Content-Type: multipart/form-data" \
     -F "file=@SampleResume.pdf"
```