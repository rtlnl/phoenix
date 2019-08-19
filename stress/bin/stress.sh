echo 'POST https://personalization-dev.rtl-di.nl/public/recommend' | \
    vegeta attack -rate 100 -duration 5s -body body.json | \
    vegeta report -type='hist[0,20ms,50ms,100ms,200ms,400ms,1s]'
