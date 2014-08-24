defmodule AtAt.Cli do

  @args """
  usage:

     at_at [ <port> ]

  start at_at on the specified port 
  by default listens on port 4040
  """	

	def main(argv) do
    argv 
    |> parse_args 
    |> process		
	end

	defp parse_args(argv) do
    parse = OptionParser.parse(argv, switches: [ help: :boolean],
                                     aliases:  [ h:    :help   ])
    case  parse  do
	    { [ help: true ],  _,  _ } -> :help
	    { _, [ port ], _     }     -> set_port(port)
	    { _, [ ],          _ }     -> set_port(4040)
	    _                          -> :help
    end
	end

  defp process(:help) do
    IO.puts(:stderr, @args)
    exit(2)
  end

  defp process(port) do
		Httpd.accept port
	end

	defp set_port(port) do
		port
	end

	
end
