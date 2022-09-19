# Pull official node base image
FROM node:18.9.0

# Set working directory
WORKDIR /app

# Add `/app/node_modules/.bin` to $PATH
ENV PATH /app/node_modules/.bin:$PATH

# Install app dependencies
COPY ui/package.json ./
COPY ui/package-lock.json ./
RUN npm install --silent

# Add app to the containter
COPY ui ./

# Start React app
CMD ["npm", "start"]