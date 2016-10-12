class Instances < Sinatra::Base
  get '/id' do
    ID
  end

  post '/session' do
    response.set_cookie 'JSESSIONID', ID
    "Please read the README.md for help on how to use sticky sessions."
  end
end
