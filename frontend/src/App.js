// Created with create-react-app (https://github.com/facebook/create-react-app)
import { 
  MemoryRouter as Router, Routes, Route, Link, useNavigate 
} from "react-router-dom"
import { useState, useEffect } from "react"
import axios from "axios"
import "./App.css"

const backendURL = "http://127.0.0.1:8080" // "https://my-site.com" 

function App() {
  return (
    <Router>
      <Routes>
        <Route path="/" element={<LogIn />} />
        <Route path="/signup" element={<SignUp />} />
        <Route path="/main" element={<Main />} />
      </Routes>
    </Router>
  )
}

function LogIn() {
  const [message, setMessage] = useState("")
  const [user, setUsername] = useState("")
  const [pass, setPassword] = useState("")
  const [tokenChecked, setTokenChecked] = useState(false)
  const navigate = useNavigate()

  const handleLogIn = async (event) => {
    event.preventDefault()
    try {
      const response = await axios.post(`${backendURL}/api/pub/login`, {
        username: user,
        password: pass,
      })
      localStorage.setItem("token", response.data.token)
      setMessage(response.data.message)
      if (response.status === 200) {
        navigate("/main")
      }
    } catch (error) {
      //console.error(error)
      setMessage(error.response.data.message)
      setUsername("")
      setPassword("")
    }
  }

  useEffect(() => {
    if (localStorage.getItem("token")) {
      setTokenChecked(true)
      navigate("/main")
    } else {
      setTokenChecked(true)
    }
  }, [navigate])

  if (!tokenChecked) {
    return <div></div>
  }
  
  return (
    <div className="container">
      <strong className="title">Web Storage</strong>
      <Link to="/signup">Sign Up </Link>
      {message && (
        <p>{message.charAt(0).toUpperCase() + message.slice(1)}</p>
      )}
      <form className="user-form" onSubmit={handleLogIn}>
          <label>Username:</label>
          <input
            type="text"
            value={user}
            onChange={(event) => setUsername(event.target.value)}
          /><br/>
          <label>Password:</label>
          <input
            type="password"
            value={pass}
            onChange={(event) => setPassword(event.target.value)}
          /><br/>
          <button type="submit">Login</button>
      </form>
    </div>
  )
}

function SignUp() {
  const [status, setStatus] = useState()
  const [message, setMessage] = useState("")
  const [user, setUsername] = useState("")
  const [pass, setPassword] = useState("")
  const [tokenChecked, setTokenChecked] = useState(false)
  const navigate = useNavigate()

  const handleSignUp = async (event) => {
    event.preventDefault()
    try {
      const response = await axios.post(`${backendURL}/api/pub/signup`, {
        username: user,
        password: pass,
      })
      setUsername("")
      setPassword("")
      setStatus(response.status)
      setMessage(response.data.message)
      if (response.status === 200) {
        setTimeout(() => {navigate("/")}, 3000)
      }
    } catch (error) {
      //console.error(error)
      setUsername("")
      setPassword("")
      setStatus(error.response.status)
      setMessage(error.response.data.message)
    }
  }

  useEffect(() => {
    if (localStorage.getItem("token")) {
      setTokenChecked(true)
      navigate("/main")
    } else {
      setTokenChecked(true)
    }
  }, [navigate])

  if (!tokenChecked) {
    return <div></div>
  }

  return (
    <div className="container">
      <strong className="title">Sign Up Page</strong>
      <Link to="/">Log In </Link>
      {message && (
        <>
          <p>{message.charAt(0).toUpperCase() + message.slice(1)}</p>
          {status === 200 && <p>Redirecting to login page in 3 seconds...</p>}
        </>
      )}
      {status !== 200 && (
        <form className="user-form" onSubmit={handleSignUp}>
            <label>Username:</label>
            <input
              type="text"
              value={user}
              onChange={(event) => setUsername(event.target.value)}
            /><br/>
            <label>Password:</label>
            <input
              type="password"
              value={pass}
              onChange={(event) => setPassword(event.target.value)}
            /><br/>
            <button type="submit">Sign Up</button>
        </form>
      )}
    </div>
  )
}

function Main() {
  const [fromDB, setFromDB] = useState([])
  const [message, setMessage] = useState("")
  const [progress, setProgress] = useState(0)
  const [editingField, setEditingField] = useState(null)
  const [editingExt, setEditingExt] = useState("")
  const [isUploading, setIsUploading] = useState(false)
  const [isDownloading, setIsDownloading] = useState(false)
  const [selectedRow, setSelectedRow] = useState(null)
  const [loading, setLoading] = useState(0)
  const [newName, setNewName] = useState("")
  const [sortOrd, setSortOrd] = useState("asc")
  const [sortCol, setSortCol] = useState("ListName")
  const navigate = useNavigate()

  const handleLogout = async () => {
    localStorage.removeItem("token")
    navigate("/")
  }

  const handleList = async () => {
    try {
      const token = localStorage.getItem("token")
      const headers = {"Authorization": `Bearer ${token}`}
      const response = await axios.get(
        `${backendURL}/api/auth/files`,
        {params: { ord: sortOrd, col: sortCol }, headers},
      )
      setFromDB(response.data.files)
      if (response.status === 200) {
        setTimeout(() => {setMessage("")}, 5000)
      }
    } catch (error) {
      //console.error(error)
      if (error.response.status === 401) {
        handleLogout()
      }
    }
  }

  const handleUpload = async (event) => {
    event.preventDefault()
    setIsUploading(true)
    const form = event.target
    const formData = new FormData(form)
    const token = localStorage.getItem("token")
    const headers = {"Authorization": `Bearer ${token}`}
    try {
      const response = await axios.post(
        `${backendURL}/api/auth/upload`, 
        formData, 
        {
          headers,
          onUploadProgress: (progressEvent) => {
            const percentage = (
              Math.round((progressEvent.loaded * 100) / progressEvent.total)
            )
            setProgress(percentage)
          }
        }
      )
      /* 
      const now = Date.now()
      const date = new Date(now)
      const formattedDate = date.toLocaleString("ru-RU", { 
        day: "numeric",
        month: "numeric",
        year: "numeric",
        hour: "numeric",
        minute: "numeric",
        second: "numeric"
      })
      console.log(formattedDate)
      console.log(response.data) 
      */
      form.reset()
      setProgress(0)
      setMessage(response.data.message)
      handleList()
    } catch (error) {
      //console.error(error)
      if (error.response.status === 401) {
        handleLogout()
      }
      setMessage(error.response.data.message)
      setProgress(0)
      form.reset()
      handleList()
    } finally {
      setIsUploading(false)
    }
  }

  const handleRename = (field) => {
    setEditingField(field.ID)
    setEditingExt(field.Extension)
    setNewName(field.ListName)
  }
  
  const handleRenameSubmit = async (event) => {
    event.preventDefault()
    try {
      const token = localStorage.getItem("token")
      const headers = {"Authorization": `Bearer ${token}`}
      await axios.post(
        `${backendURL}/api/auth/rename`, 
        {id: editingField, name: newName, extension: editingExt}, 
        {headers},
      )
      setEditingField(null)
      setNewName("")
      handleList()
    } catch (error) {
      //console.error(error)
      if (error.response.status === 401) {
        handleLogout()
      }
      setMessage(error.response.data.message)
      handleList()
    }
  }

  const handleDownload = async (value) => {
    try {
      setIsDownloading(true)
      setSelectedRow(value[0])
      const token = localStorage.getItem("token")
      const headers = {"Authorization": `Bearer ${token}`}
      const response = await axios.get(
        `${backendURL}/api/auth/download`,
        {params: { id: value[0], file: value[1] }, 
          headers, 
          responseType: "blob",
          onDownloadProgress: (progressEvent) => {
            const percentCompleted = Math.round(
              (progressEvent.loaded * 100) / progressEvent.total
            )
            console.log(`Progress: ${percentCompleted}%`)
            setLoading(percentCompleted)
          }
        }
      )
      const url = window.URL.createObjectURL(new Blob([response.data]))
      const link = document.createElement("a")
      link.href = url
      link.setAttribute("download", value[1])
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)
    } catch (error) {
      //console.error(error)
      if (error.response.status === 401) {
        handleLogout()
      }
      setMessage("An error occurred while preparing the file.")
    } finally {
      setIsDownloading(false)
      setSelectedRow(null)
    }
  }

  const handleDelete = async (event) => {
    const value = parseInt(event.target.value, 10)
    try {
      const token = localStorage.getItem("token")
      const headers = {"Authorization": `Bearer ${token}`}
      await axios.post(`${backendURL}/api/auth/delete`, {id: value}, {headers})
      handleList()
    } catch (error) {
      //console.error(error)
      if (error.response.status === 401) {
        handleLogout()
      }
    }
  }

  const handleSort = (column) => {
    setSortCol(column)
    if (sortCol === column) {
      setSortOrd(sortOrd === "asc" ? "desc" : "asc")
    } else {
      setSortCol(column)
      setSortOrd("asc")
    }
  }

  const formatDate = (date) => {
    const options = {
      day: "2-digit",
      month: "2-digit",
      year: "numeric",
      hour: "2-digit",
      minute: "2-digit",
      hour12: false,
      timeZone: Intl.DateTimeFormat().resolvedOptions().timeZone,
    }
    const formattedDate = new Intl.DateTimeFormat(
      undefined, 
      options
    ).format(new Date(date))
    return formattedDate
  }

  function formatSize(size) {
    if (size < 1024000) {
      return `${(size / 1024).toFixed(0)} KB`
    } else if (size < 1024000000) {
      return `${(size / 1024000).toFixed(2)} MB`
    } else {
      return `${(size / 1024000000).toFixed(2)} GB`
    }
  }

  useEffect(() => {
    handleList()
    // eslint-disable-next-line
  }, [sortCol, sortOrd])

  return (
    <div className="container-main">
      <h3>Main Page</h3>
      <button className="logout-button" onClick={handleLogout}>Logout</button>
      <br/>
      <form onSubmit={handleUpload} className="upload-form">
        <input className="upload-input" type="file" name="files" multiple/>
        <button className="upload-button" type="submit" disabled={isUploading}>
          {isUploading ? "Uploading..." : "Upload"}
        </button>
      </form>
      {progress > 0 && <p>Uploading: {progress}%</p>}
      {message && (
        <p>{message.charAt(0).toUpperCase() + message.slice(1)}</p>
      )}
      <table className="data-table">
        <thead>
          <tr>
            <th className="name-column" onClick={() => handleSort("ListName")}>
              Name 
              {sortCol === "ListName" && sortOrd === "asc" && "▲"}
              {sortCol === "ListName" && sortOrd === "desc" && "▼"}
            </th>
            <th className="ext-column" onClick={() => handleSort("Extension")}>
              Ext
              {sortCol === "Extension" && sortOrd === "asc" && "▲"}
              {sortCol === "Extension" && sortOrd === "desc" && "▼"}
            </th>
            <th className="date-column" onClick={() => handleSort("Date")}>
              Date 
              {sortCol === "Date" && sortOrd === "asc" && "▲"}
              {sortCol === "Date" && sortOrd === "desc" && "▼"}
            </th>
            <th className="size-column" onClick={() => handleSort("Size")}>
              Size 
              {sortCol === "Size" && sortOrd === "asc" && "▲"}
              {sortCol === "Size" && sortOrd === "desc" && "▼"}
            </th>
            <th className="action-column">
              Actions
            </th>
          </tr>
        </thead>
        <tbody>
          {fromDB.map((field) => (
            <tr key={field.ID}>
              <td className="name-column">
                {editingField === field.ID ? (
                  <form onSubmit={handleRenameSubmit}> 
                    <input
                      type="text"
                      value={newName}
                      onChange={(event) => setNewName(event.target.value)}
                    />
                    <button type="submit">Save</button>
                  </form>
                ) : (
                  field.ListName
                )}
              </td>
              <td className="ext-column">{field.Extension}</td>
              <td className="date-column">{formatDate(field.Date)}</td>
              <td className="size-column">{formatSize(field.Size)}</td>
              <td className="action-column">
                {editingField === field.ID ? (
                  <button disabled>Rename</button>
                ) : (
                  <button onClick={() => handleRename(field)}>Rename</button>
                )}
                <button
                  type="button"
                  onClick={() => handleDownload([field.ID, field.ListName + field.Extension])}
                  disabled={isDownloading || selectedRow === field.ID}
                >
                  {isDownloading && selectedRow === field.ID ? `${loading}%` : "Download"}
                </button>
                
                <button type="submit" value={field.ID} onClick={handleDelete}>
                  Delete
                </button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}


export default App
