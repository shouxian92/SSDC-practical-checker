![build](https://github.com/shouxian92/SSDC-practical-checker/workflows/build/badge.svg)

# Description

Singapore has a ridiculous amount of people who wants to learn driving each day. It is not possible for a single school to accomodate everyone to book their driving lessons. Driving lessons are difficult to book if you want a lesson in the near future (e.g. this week). On average, a person waits at least a month before starting their practical driving lessons.

I can't spend 20 seconds to type my username and password into a website that expires my ASP.NET session almost every 30 minutes. So i spent a few hours of my life writing something that can help me to do some web crawling so that I'm able to get notified if there is a driving lesson available for booking.

# Requirements

* [Golang](https://golang.org/dl/) (1.14)

# How to use

1. Fill in your username and password in the `.credentials.template.json` file
2. Rename it to `.credentials.json`
3. Run the script with the following command

```properties
go run main.go
```

4. Script will run at a fixed interval

# Todos

1. Send notifications/email if there is an available lesson this week
2. Support input parameter for the fixed polling interval instead of hardcoding it
3. Increase search range for booking (beyond the current week)