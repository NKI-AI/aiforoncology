def main():
    import sys
    from tensorboard import program

    # Create a new argv list with "tensorboard" as the first element
    # followed by any command-line arguments passed to this script
    argv = ["tensorboard"] + sys.argv[1:]

    tb = program.TensorBoard()
    tb.configure(argv=argv)
    tb.main()


if __name__ == "__main__":
    main()
