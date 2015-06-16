FROM golang:onbuild

CMD ["app", "-logtostderr", "-stderrthreshold=INFO"]
