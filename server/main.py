import http.server
import multiprocessing
import os
import signal
import sys

PORTS = [5000, 6969, 7070]
DIRECTORY = os.path.dirname(os.path.abspath(__file__))


def run_server(port: int) -> None:
    handler = http.server.SimpleHTTPRequestHandler
    server = http.server.HTTPServer(("", port), handler)
    os.chdir(DIRECTORY)
    print(f"[+] Server running at http://localhost:{port}")
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        pass
    finally:
        server.server_close()
        print(f"[-] Server on port {port} stopped")


def main() -> None:
    processes = [multiprocessing.Process(target=run_server, args=(port,), daemon=True) for port in PORTS]

    for p in processes:
        p.start()

    print(f"Serving '{DIRECTORY}' on ports: {PORTS}")
    print("Press Ctrl+C to stop all servers.\n")

    def shutdown(sig, frame):
        print("\nShutting down all servers...")
        for p in processes:
            p.terminate()
        sys.exit(0)

    signal.signal(signal.SIGINT, shutdown)
    signal.signal(signal.SIGTERM, shutdown)

    for p in processes:
        p.join()


if __name__ == "__main__":
    main()
