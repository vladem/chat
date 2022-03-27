import docker
import pytest
from typing import Optional, Any, List
import time
import subprocess as sp
import random
import string
import os


class ServerContainer:
    server_image_name = "chat-server" 

    def __init__(self):
        self.container: Optional[Any] = None

    def __enter__(self):
        client = docker.from_env()
        self.container = client.containers.run(self.server_image_name, network_mode="container:chat_tests", detach=True)
        return self

    def __exit__(self, *args, **kwargs):
        print("server stderr:\n" + self.container.logs().decode())
        self.container.stop()


@pytest.fixture
def server():
    return ServerContainer()


def random_str():
    return ''.join(random.choices(string.ascii_uppercase + string.digits, k=10))


def send_messages(sender: sp.Popen) -> List[str]:
    msgs = [random_str() for _ in range(10)]
    for msg in msgs:
        sender.stdin.write(f"{msg}\n".encode())
        print("sent: " + msg)
    sender.stdin.flush()
    return msgs


def readline_with_timeout(pipe, timeout_s = 1) -> str:
    line = ""
    start = time.time()
    while not line and (time.time() - start < timeout_s):
        line = pipe.readline()
        time.sleep(0.01)
    assert line, "timeout on waiting for new line"
    return line.decode()


def wait_messages(proc: sp.Popen, msgs: List[str]):
    while msgs:
        line = readline_with_timeout(proc.stdout)
        if msgs[0] not in line:
            continue
        print("received: " + line)
        msgs = msgs[1:]


def send_receive():
    user1 = random_str()
    user2 = random_str()
    sender_cmd = ["/client", "--receiver_id", user1, "--sender_id", user2]
    receiver_cmd = ["/client", "--receiver_id", user2, "--sender_id", user1]
    sender, receiver = None, None
    try:
        print("running: " + str(receiver_cmd))
        receiver = sp.Popen(receiver_cmd, stdin=sp.PIPE, stdout=sp.PIPE, stderr=sp.PIPE)
        os.set_blocking(receiver.stdout.fileno(), False)
        print("running: " + str(sender_cmd))
        sender = sp.Popen(sender_cmd, stdin=sp.PIPE, stdout=sp.PIPE, stderr=sp.PIPE)
        msgs = send_messages(sender)
        wait_messages(receiver, msgs)
    finally:
        if receiver:
            receiver.kill()
        if sender:
            sender.kill()


def test_many_chats(server):
    with server:
        for i in range(5):
            send_receive()
            print("done iter i = " + str(i))
            # todo: check, why server stops responding after first chat is closed))
        # todo: check memory consumption
