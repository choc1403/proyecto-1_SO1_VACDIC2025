import time

if __name__ == '__main__':
    data = []
    
    while True:
        data.append("A"*10_000_000)
        time.sleep(60)