import json
import web3
import time
w3=web3.Web3(web3.HTTPProvider('http://127.0.0.1:8545', request_kwargs={'timeout': 60 * 10}))

contract_abi = open("./compile/contract/Email.abi",'r').read()
# "CONTRACT_ADDRES"



import time
from watchdog.observers import Observer
from watchdog.events import FileSystemEventHandler

# 继承 FileSystemEventHandler 并重写 on_modified 方法
class MyHandler(FileSystemEventHandler):
    def on_modified(self, event):
        print(f'read new contract addr from {event.src_path} ')
        contract_address=open("./compile/contract/addr.txt",'r').read()
        contract = w3.eth.contract(address=contract_address, abi=contract_abi)
        message_event = contract.events.Event()
        block_filter = w3.eth.filter({'fromBlock': 50, 'address': contract_address})

        while True:
            entries = block_filter.get_new_entries()
            for entry in entries:
                # print(f"block_filter_length: {len(entries)}")
                receipt = w3.eth.wait_for_transaction_receipt(entry['transactionHash'])
                result = message_event.processReceipt(receipt)
                
                obj=result[0].args
                res="event:"+obj.eventName+", sender"+obj.sender+", value:"+str(obj.value)+", cid:"+obj.cid+", extra:"+str(obj.extra)+"\n"
                print("new event ", res)
#                 open("a.txt","a").write(res)
            time.sleep(1)
        

if __name__ == "__main__":
    path = './compile/contract/addr.txt'
    observer = Observer()
    event_handler = MyHandler()
    observer.schedule(event_handler, path, recursive=True)
    observer.start()
    try:
        while True:
            time.sleep(1)
    except KeyboardInterrupt:
        observer.stop()
    observer.join()






