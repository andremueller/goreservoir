library(tidyr)
library(dplyr)
library(readr)
library(ggplot2)
result <- read_csv("../cmd/health/result.csv") %>% as_tibble()
p <- result %>% ggplot(aes(x=index,y=ks.stat)) + geom_line()

print(p)