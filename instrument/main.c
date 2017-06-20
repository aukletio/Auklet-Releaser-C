int f(int n)
{
	return n + 1;
}

int
main(int argc, char **argv)
{
	for (int i = 0; i < 30; i++) {
		f(i);
	}
	return 0;
}
