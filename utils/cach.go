package utils

import "fmt"

type RequestAnswer struct {
	Name   string
	RType  uint16
	R      bool
	Answer []byte
}

type identityKey struct {
	Name  string
	RType uint16
}

type Cach struct {
	Size  int
	Queue Queue[identityKey]
	Names map[identityKey]*RequestAnswer
}

func (c Cach) Get(name string, rType uint16) ([]byte, error) {
	key := createIdentityKey(name, rType)
	if answer, ok := c.Names[key]; ok {
		answer.R = true
		return answer.Answer, nil
	}
	return nil, fmt.Errorf("данного запроса нет в кеше")
}

func createIdentityKey(name string, rType uint16) identityKey {
	return identityKey{
		Name:  name,
		RType: rType,
	}
}

// def contains(self, url):
//     return url in self.pages

// def put(self, url, page_code, response_code):
//     while self.queue.full():
//         page = self.queue.get()
//         if page.r:
//             page.r = False
//             self.queue.put(page)
//         else:
//             self.pages.pop(page.url)
//     page = Page(url, page_code, response_code)
//     self.pages[url] = page
//     self.queue.put(page)
