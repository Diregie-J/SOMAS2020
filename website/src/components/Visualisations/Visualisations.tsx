import React, { useEffect, useState } from 'react'
import { Button, Alert } from 'react-bootstrap'
import { useHistory, Route, Switch } from 'react-router-dom'
import VisualisationsNavbar from './VisualisationsNavbar'
import {
  gamevisualisation,
  visualisations,
  iifovisualisation,
  iigovisualisation,
  iitovisualisation,
  resourcesvisualisation,
  rolesvisualisation,
} from '../../consts/paths'
import { OutputJSONType } from '../../consts/types'
import { GitHash } from '../../consts/info'
import {
  initialLoadingState,
  useLoadingState,
} from '../../contexts/loadingState'
import {
  loadLocalVisOutput,
  clearLocalVisOutput,
  storeLocalVisOutput,
} from './utils'
import Game from './Game/Game'
import IIFO from './IIFO/IIFO'
import IITO from './IITO/IITO'
import IIGO from './IIGO/IIGO'
import Resources from './Resources/Resources'
import Roles from './Roles/Roles'
import FourOhFour from '../FourOhFour/FourOhFour'

const Visualisations = () => {
  const [output, setOutput] = useState<OutputJSONType | undefined>(undefined)
  const [, setLoading] = useLoadingState()
  const history = useHistory()
  const [error, setError] = useState<string | undefined>(undefined)
  const [warning, setWarning] = useState<string | undefined>(undefined)

  const onDidMount = async () => {
    window.scrollTo(0, 0)
    setLoading({ loading: true, loadingText: `We're hard at work!` })
    try {
      const o = await loadLocalVisOutput()
      if (o) {
        setOutput(o)
        history.push(gamevisualisation)
      }
    } catch (err) {
      // if error, just assume not stored at all
      console.error(err)
    }
    setLoading(initialLoadingState)
  }

  useEffect(() => {
    onDidMount()
  }, [])

  const handleReset = async () => {
    setLoading({ loading: true, loadingText: `Cleaning up your mess!` })
    setOutput(undefined)
    await clearLocalVisOutput()
    history.push(visualisations)
    setLoading(initialLoadingState)
    setError(undefined)
    setWarning(undefined)
  }

  useEffect(() => {
    if (output) {
      try {
        const gotGitHash = output.GitInfo.Hash
        if (gotGitHash !== GitHash) {
          setWarning(
            `This website was built on commit "${GitHash}", and the output you're trying to visualise ` +
              ` was produced on commit "${output.GitInfo.Hash}". There may be incompatibilities!`
          )
        } else {
          setWarning(undefined)
        }
      } catch (err) {
        // can't read output.GitInfo.Hash, just reset which clears localforage as well
        handleReset()
      }
    }
  }, [output])

  const onUpload = async (event: React.ChangeEvent<HTMLInputElement>) => {
    setLoading({ loading: true, loadingText: `Reading your file!` })

    try {
      const { files } = event.target
      const file: Blob | null = files !== null ? files[0] : null
      if (!file || !(file instanceof Blob)) {
        throw new Error('No or unsupported file uploaded')
      }
      const outputText = await file.text()
      const o = JSON.parse(outputText) as OutputJSONType
      try {
        // find githash to check whether the JSON is ok
        const gotGitHash = o.GitInfo.Hash
        console.debug(gotGitHash)
      } catch (err) {
        throw new Error(`Unsupported file.`)
      }
      setOutput(o)
      await storeLocalVisOutput(o)
      setError(undefined)
    } catch (err) {
      setError(err.message)
    }
    history.push(gamevisualisation)
    setLoading(initialLoadingState)
  }

  return (
    <>
      {output && (
        <>
          <VisualisationsNavbar reset={handleReset} />
        </>
      )}
      <div style={{ paddingTop: 24 }}>
        {error && (
          <Alert
            variant="danger"
            onClose={() => setError(undefined)}
            dismissible
            className="custom-alert"
          >
            <Alert.Heading>Oh reeeeeeeeee!</Alert.Heading>
            <p>{error}</p>
          </Alert>
        )}
        {warning && (
          <Alert
            variant="warning"
            onClose={() => setWarning(undefined)}
            dismissible
            className="custom-alert"
          >
            <Alert.Heading>Rough seas ahead!</Alert.Heading>
            <p>{warning}</p>
          </Alert>
        )}
        {output ? (
          <Switch>
            <Route
              path={gamevisualisation}
              exact
              component={() => <Game output={output} />}
            />
            <Route
              path={iigovisualisation}
              exact
              component={() => <IIGO output={output} />}
            />
            <Route
              path={iitovisualisation}
              exact
              component={() => <IITO output={output} />}
            />
            <Route
              path={iifovisualisation}
              exact
              component={() => <IIFO output={output} />}
            />
            <Route
              path={rolesvisualisation}
              exact
              component={() => <Roles output={output} />}
            />
            <Route
              path={resourcesvisualisation}
              exact
              component={() => <Resources output={output} />}
            />
            <Route component={FourOhFour} />
          </Switch>
        ) : (
          <>
            <h1>Visualisations</h1>
            <h5 style={{ marginTop: 24 }}>Upload output JSON file</h5>

            <Button variant="warning">
              <label htmlFor="multi" style={{ margin: 0 }}>
                Upload
              </label>
              <input
                style={{ display: 'none' }}
                type="file"
                accept=".json"
                id="multi"
                onChange={onUpload}
              />
            </Button>
          </>
        )}
      </div>
    </>
  )
}

export default Visualisations
