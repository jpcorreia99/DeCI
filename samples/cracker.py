
import hashlib
import hashlib
import multiprocessing
import time
import itertools
import uuid
import hashlib
import sys


hash, salt = sys.argv[1], sys.argv[2]
def block_func(combination):
    """Take one combination, compute its hash, and return if a match is found"""
    password = "".join(combination)
    h = hashlib.sha256(salt.encode() + password.encode())
    digest = h.hexdigest()
    if digest == hash:
        return password

charset = "0123456789"


def find():
    # create a pool of processes
    with multiprocessing.Pool(processes=1) as pool:
        length = 7
        # create the set of passwords for this length
        combinations_generator = itertools.product(charset, repeat=length)

        # compute number of possible passwords
        total_combinations = len(charset)**length

        #read the generator lazily and map combinations to the function
        for i, foundPw in enumerate(pool.imap_unordered(block_func, combinations_generator, 1000)):
            if foundPw is not None:
                return foundPw

print(find())