beanstalk frontend
ec2 backend
lambda for temperature calls
dynamo for storing data(10s)
ses for sending emails
optional gh repo for app

alert 8001
socket 8080
stats 8000
website 5173/80

alert uses alert db

socket and stats use the same ./temperature.db
