import { submitAction } from './actions'

export default function Home() {
  return (
    <div>
      <h1>ShellyGo Lab - Vulnerable App</h1>
      <form action={submitAction}>
        <input type="text" name="input" placeholder="Type something..." />
        <button type="submit">Submit</button>
      </form>
    </div>
  )
}
