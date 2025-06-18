#!/usr/bin/env python3


def main():
    from tensorboard import program

    tb = program.TensorBoard()
    tb.configure(argv=[None, "--logdir=./logs"])
    tb.main()


if __name__ == "__main__":
    main()
