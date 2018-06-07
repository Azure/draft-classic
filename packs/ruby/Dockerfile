FROM ruby
ENV PORT 3000
EXPOSE 3000
RUN bundle config --global frozen 1

WORKDIR /usr/src/app

COPY Gemfile Gemfile.lock ./
RUN bundle install

COPY . .
CMD ["ruby", "app.rb"]
