import json
import web3
import time
w3=web3.Web3(web3.HTTPProvider('http://127.0.0.1:8545', request_kwargs={'timeout': 60 * 10}))

contract_abi = open("./compile/contract/Email.abi",'r').read()

# import os
# os.system("pip3 install py-solc-x==2.0.2")
# os.system("pip3 install web3==6.15.1")

import time
from watchdog.observers import Observer
from watchdog.events import FileSystemEventHandler
import threading

# 继承 FileSystemEventHandler 并重写 on_modified 方法
class MyHandler(FileSystemEventHandler):
    def __init__(self):
        super(MyHandler, self).__init__()
        # self.lock = threading.Lock()
        self.cnt=0#合约重新部署的次数

    def on_modified(self, event):
        self.cnt+=1
        print(f'read new contract addr from {event.src_path} ')
        contract_address=open(event.src_path,'r').read()
        contract = w3.eth.contract(address=contract_address, abi=contract_abi)
        self.message_event = contract.events.Event()
        self.block_filter = w3.eth.filter({'fromBlock': 1, 'address': contract_address})

        threading.Thread(target=self.process_event, args=(event,)).start()

    def process_event(self, event):       
        currentCNT = self.cnt
        while currentCNT == self.cnt:
            entries = self.block_filter.get_new_entries()
            for entry in entries:
                # print(f"block_filter_length: {len(entries)}")
                receipt = w3.eth.wait_for_transaction_receipt(entry['transactionHash'])
                # print(dir(self.message_event))
                result = self.message_event.process_receipt(receipt)
                
                for i in range(0, len(result)):
                    obj=result[i].args
                    res="event:"+obj.eventName+", sender:"+obj.sender+", value:"+str(obj.value)+", field:"+obj.fid+", extra:"+str(obj.extra)+"\n"
                    print(res)
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






