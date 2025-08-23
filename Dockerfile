FROM python:3.12-alpine

WORKDIR /app

RUN pip install docker

COPY . .

ENV PORT=4531

EXPOSE ${PORT}

CMD ["python", "app.py"]