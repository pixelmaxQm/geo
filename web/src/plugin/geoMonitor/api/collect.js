import service from '@/utils/request'

export const runCollection = (data) => {
  return service({
    url: '/geoMonitor/collect/run',
    method: 'post',
    data
  })
}

export const getCollectionTask = (id) => {
  return service({
    url: `/geoMonitor/collect/task/${id}`,
    method: 'get'
  })
}

export const getCollectionResultList = (params) => {
  return service({
    url: '/geoMonitor/collect/result/list',
    method: 'get',
    params
  })
}
