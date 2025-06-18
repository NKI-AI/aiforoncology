from vllm.entrypoints.openai.api_server import (
    FlexibleArgumentParser,
    make_arg_parser,
    validate_parsed_serve_args,
    uvloop,
    run_server,
)


def main():
    """Main entry point for the API server."""
    parser = FlexibleArgumentParser(description="vLLM OpenAI-Compatible RESTful API server")
    parser = make_arg_parser(parser)
    args = parser.parse_args()
    validate_parsed_serve_args(args)
    uvloop.run(run_server(args))


if __name__ == "__main__":
    main()
