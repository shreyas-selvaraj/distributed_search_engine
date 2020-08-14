# distributed_search_engine
Distributed search engine written in golang deployed using Docker and Kubernetes on GCP. Concurrently scrapes the web from a starting link using goroutines and uploads the links along with five tags determined using the Rapid Automatic Keyword Extraction (RAKE) algorithm to ElasticSearch. Results for a search query are then grabbed using ElasticSearch. 
