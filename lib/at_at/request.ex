defmodule AtAt.Request do

	def parse([]), do: []
	def parse(req), do: _parse(req |> String.split("\n") |> Enum.map(&String.strip/1))

	# set defaults
	defp _parse(list, acc \\ %{})

	defp _parse([], result), do: result
	
	# two empty lines means we are done with headers and the body (may) start
	defp _parse(["", "" | body], result) do		
		body = body |> Enum.join "\n"
		_parse([], _merge_results(:body, body, result) )
	end

	defp _parse([line | rest], result) do
		line = line |> String.strip

		cond do
			# if request is valid, this one is executed only one time
			line =~ ~r/^(GET|POST|PUT|DELETE|HEAD|TRACE|OPTIONS)/i ->
				[method, path, version] = line |> String.split
				_parse(rest, %{method: method |> to_sym, path: path, version: version})

			line =~ ~r/^[^:]+:\s?(.*)$/ -> 
				headers = _parse_header(line) |> _merge_results result[:headers]
				_parse(rest, _merge_results(:headers, headers, result))

			true -> %{error: "invalid request"}
		end

	end

	defp _parse_header(header) do 
		[name, value] = String.split(header, ":", parts: 2) |> Enum.map(&String.strip/1)
		[name |> to_sym, value]
	end

	defp _merge_results(name, value, results) when is_map(results) do		
		results |> Map.put name, value
	end

	defp _merge_results([name, value], results) when is_map(results) do		
		_merge_results name, value, results
	end

	defp _merge_results(kwarg, _results), do: _merge_results(kwarg, %{})

	defp to_sym(s), do: s |> String.downcase |> String.to_atom

end