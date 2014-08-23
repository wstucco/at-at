defmodule RequestTest do
  use ExUnit.Case

  alias AtAt.Request

  test "invalid requests are rejected" do
	  invalid_req = """
		GET  /blog/2014/02/12/pattern-matching-and-parsing/ HTTP/1.1
		Host: massimoronca.dev
		Content-Type: text/plain

		body line #1
		body line #2  
	  """

  	assert AtAt.Request.parse(invalid_req) == %{error: "invalid request"}
	end

  test "requests methods are accorectly recognized" do
	  get_req    = "GET / HTTP/1.1"
	  post_req   = get_req |> String.replace("GET", "POST")
	  put_req    = get_req |> String.replace("GET", "PUT")
	  delete_req = get_req |> String.replace("GET", "DELETE")

	  get_res    = %{method: :get    , path: "/", version: "HTTP/1.1"}
	  post_res   = %{method: :post   , path: "/", version: "HTTP/1.1"}
	  put_res    = %{method: :put    , path: "/", version: "HTTP/1.1"}
	  delete_res = %{method: :delete , path: "/", version: "HTTP/1.1"}

  	assert AtAt.Request.parse(get_req)    == get_res
  	assert AtAt.Request.parse(post_req)   == post_res
  	assert AtAt.Request.parse(put_req)    == put_res
  	assert AtAt.Request.parse(delete_req) == delete_res
	end	

	test "complete requests are correctly recognized" do

	  get_req = """
		GET  / HTTP/1.1
		Host: massimoronca.dev
		Content-Type: text/plain


		body line #1
		body line #2
	  """

	  get_res = %{
	  	method: :get,
	  	path: "/",
	  	version: "HTTP/1.1",
	  	headers: %{"content-type": "text/plain", host: "massimoronca.dev"},
	  	body: "body line #1\nbody line #2\n"
	  }

		assert AtAt.Request.parse(get_req) == get_res	  
	end

end
