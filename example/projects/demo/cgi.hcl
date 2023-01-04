lambda "echo" {
  exec = ["cat", "-"]
}

lambda "date" {
  exec = ["./time.sh"]
}

post "echo" {
  call "echo" {}
}

get "date" {
  call "date" {
    environment = {
      FORMAT = "{{.Query.Get \"format\"}}" // use %25 to escape % in query
    }
  }
}