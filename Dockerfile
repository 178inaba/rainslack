FROM golang:1.4.2-onbuild

CMD ["app", "-logtostderr", "-stderrthreshold=INFO"]
