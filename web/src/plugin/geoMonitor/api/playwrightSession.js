import service from '@/utils/request'

export const startPlaywrightSession = (data) => {
  return service({
    url: '/geoMonitor/playwright/session/start',
    method: 'post',
    data
  })
}

export const getPlaywrightSession = (id) => {
  return service({
    url: `/geoMonitor/playwright/session/${id}`,
    method: 'get'
  })
}

export const refreshPlaywrightSession = (id) => {
  return service({
    url: `/geoMonitor/playwright/session/${id}/refresh`,
    method: 'post'
  })
}

export const deletePlaywrightSession = (id) => {
  return service({
    url: `/geoMonitor/playwright/session/${id}`,
    method: 'delete'
  })
}
