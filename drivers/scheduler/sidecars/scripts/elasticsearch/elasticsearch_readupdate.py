from elasticsearch import Elasticsearch
import argparse
import time
import random
import sys
import pdb
import string
import copy
query = {
  "query": {"match_all": {}}
 }
source_to_update = {}
elasticsearch_user_name ='es_username'
elasticsearch_user_password ='es_password'

# Just to control the minimum value globally (though its not configurable)
def generate_random_int(max_size):
    try:
        return random.randint(1, max_size)
    except:
        print("Not supporting {0} as valid sizes!".format(max_size))
        sys.exit(1)

def modified_source(source_doc):
    source_doc_cp = dict()
    for key in source_doc.keys():
        source_doc_cp[key] = ''.join(generate_random_string(255))
    return source_doc_cp
    
def generate_random_string(max_size):
    return ''.join(random.choice(string.ascii_lowercase) for _ in range(generate_random_int(max_size)))

def main():
    # Set a parser object
    parser = argparse.ArgumentParser()
    parser.add_argument("--es_address", nargs='+', help="The address of your cluster (no protocol or port)", required=True)
    args = parser.parse_args()
    #args.es_address
    #es= Elasticsearch("esnode-0.elasticsearch-api.elasticsearchrepl1.svc.cluster.local:9200", http_auth=None, verify_certs=True,  ssl_context= None,  timeout=1200)
    es= Elasticsearch(args.es_address, http_auth=None, verify_certs=True,  ssl_context= None,  timeout=1200)
    #For read evey index and later update 1/5 document  inside each data.
    for es_index in es.indices.get('*'):
        print("es index "+es_index)
        #using scroll API to search docs for given indice
        result = es.search(index=es_index, body={"query":{"match_all":{}}}, size=100 , scroll="5m")
        for doc in result['hits']['hits']:
            print ("DOC ID"+doc['_id'])#, doc['_source']
        old_scroll_id = result['_scroll_id']
        results = result['hits']['hits']
        print("Lenght of hits "+str(len(results)))
        try:
            counter = 0
            while len(results)>0:
                for i, r in enumerate(results):
                    counter+=1
                    print(i)
                    print r['_id']
                    if counter == 5:
                        #print(r)
                        #add_fields()
                        source_doc_cp = modified_source(r['_source'])
                        try:
                            response = es.update(index=es_index,doc_type="stresstest", id=r['_id'], body={"doc":source_doc_cp},refresh=True)
                            print('Update doc response :', response)
                        except Exception as ex:
                            print("Exception while updating %s : %s", es_index, ex)
                        counter = 0
                        #response = es.update(index=es_index, doc_type="stresstest", id=r['_id'], body=add_fields())
                result = es.scroll(scroll_id=old_scroll_id,
                    scroll='5m'  # length of time to keep search context
                    )
                # check if there's a new scroll ID
                if old_scroll_id != result['_scroll_id']:
                    print("NEW SCROLL ID:", result['_scroll_id'])
                # keep track of pass scroll _id
                old_scroll_id = result['_scroll_id']
                results = result['hits']['hits']
                print("Scroll id "+old_scroll_id)
                print("Lenght of hits " +str(len(results)))
        except Exception as ex:
            print("Exception thrown while querying %s : %s",es_index, ex)
        print("Picking next indes after 1 minute")
        time.sleep(60)    
try:
    main()
except Exception as e:
    print("Got unexpected exception. probably a bug, please report it.")
    print("")
    print(e.message)
    print("")
    sys.exit(1)