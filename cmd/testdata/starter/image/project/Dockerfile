# Dockerfile for building your application.
# Defines the final image that contains content from both the image and template.

FROM registry.access.redhat.com/ubi7/ubi

WORKDIR /project

COPY . ./

EXPOSE 8080

# Pass control your application
CMD ["/bin/bash",  "/project/userapp/hello.sh"]