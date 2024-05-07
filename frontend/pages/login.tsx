import React, { useState , useEffect } from 'react';
import { useRouter } from 'next/router';

const LoginPage: React.FC = () => {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [otp, setOtp] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [isRegistering, setIsRegistering] = useState(false);
  const [isForgetting, setIsForgetting] = useState(false);
  const [otpSentTime, setOtpSentTime] = useState<number | null>(null);
  const [remainingTime, setRemainingTime] = useState<number | null>(null);
  const router = useRouter();

  useEffect(() => {
    let interval: NodeJS.Timeout;
    if (otpSentTime !== null) {
      // Start a timer to update the remaining time
      interval = setInterval(() => {
        const currentTime = Date.now();
        const elapsedTime = currentTime - otpSentTime;
        const timeLeft = Math.max(0, 90 - Math.floor(elapsedTime / 1000)); // Calculate remaining time in seconds
        setRemainingTime(timeLeft);
        if (timeLeft === 0) {
          // Stop the timer if 90 seconds have passed
          clearInterval(interval);
        }
      }, 1000);
    }
    return () => clearInterval(interval);
  }, [otpSentTime]);
  

  const handleRegister = async () => {
    setLoading(true);
    setError(null);
  
    try {
      // Proceed with registration
      const response = await fetch('https://edu-wise-backend.onrender.com/api/register', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ username: username + '@iitk.ac.in', password }),
      });
  
      if (response.ok) {
        setIsRegistering(true); // Show OTP field after successful registration
        alert('OTP sent successfully'); // Set success message
        setError(null); // Clear any previous errors
        setOtpSentTime(Date.now()); // Record the time when OTP is sent
      } else {
        const errorData = await response.json(); // Fetch error message from the backend
        setError(errorData.error || 'Registration failed. Please try again.');
      }
    } catch (error) {
      setError('An error occurred. Please try again.');
    }
  
    setLoading(false);
  };
  
  
  
  const handleVerifyOTP = async () => {
    setLoading(true);
    setError(null);
  
    try {
      const response = await fetch('https://edu-wise-backend.onrender.com/api/verify', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ username: username + '@iitk.ac.in', otp }),
      });
  
      if (response.ok) {
        alert('OTP verified successfully'); // Set success message
        router.push('/login'); // Redirect to login page after OTP verification
      } else {
        const errorMessage = await response.text();
        setError(errorMessage || 'OTP verification failed. Please try again.');
      }
    } catch (error) {
      setError('An error occurred. Please try again.');
    }
  
    setLoading(false);
  };

  const handleForgetPassword = async () => {
    setLoading(true);
    setError(null);
  
    try {
      const response = await fetch('https://edu-wise-backend.onrender.com/api/forget', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          username: username + '@iitk.ac.in',
          newPassword: newPassword, // Include the newPassword field
        }),
      });
  
      if (response.ok) {
        const responseData = await response.json();
        alert(responseData.message); // Display success message
      } else {
        const errorMessage = await response.text();
        setError(errorMessage || 'Password reset failed. Please try again.');
      }
    } catch (error) {
      setError('An error occurred. Please try again.');
    }
  
    setLoading(false);
  };

  

  const handleLogin = async () => {
    setLoading(true);
    setError(null);
  
    try {
      const response = await fetch('https://edu-wise-backend.onrender.com/api/login', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ username: username + '@iitk.ac.in', password }),
      });
  
      if (response.ok) {
        const data = await response.json();
        localStorage.setItem('token', data.token); // Set JWT token in localStorage
        router.push('/main');
      } else {
        const errorMessage = await response.text();
        setError(errorMessage || 'Login failed. Please try again.');
      }
    } catch (error) {
      setError('An error occurred. Please try again.');
    }
  
    setLoading(false);
  };

  const handleResendOTP = async () => {
    setLoading(true);
    setError(null);

    try {
      const response = await fetch('https://edu-wise-backend.onrender.com/api/resend', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ username: username + '@iitk.ac.in' }),
      });

      if (response.ok) {
        alert('OTP resent successfully');
        setError(null);
        setOtpSentTime(Date.now()); // Record the time when OTP is resent
      } else {
        const errorMessage = await response.text();
        setError(errorMessage || 'Failed to resend OTP. Please try again.');
      }
    } catch (error) {
      setError('An error occurred. Please try again.');
    }

    setLoading(false);
  };

  return (
    <div className="flex flex-col items-center justify-center min-h-screen">
      <h1 className="text-5xl text-white font-bold mb-4">EduWise</h1>
      <div className="bg-gray-300 rounded-2xl p-6 mb-4">
        <div className="flex flex-col items-center">
        {isRegistering && !isForgetting && (
          <h1 className="text-3xl text-black mb-4">Register Page</h1>
        )}
        {!isRegistering && !isForgetting && (
          <h1 className="text-3xl text-black mb-4">Login Page</h1>
        )}
        {!isRegistering && isForgetting && (
          <h1 className="text-3xl text-black mb-4">Forget Password Page</h1>
        )}
          <div className="mb-2">
            <input
              type="text"
              placeholder="IITK Username"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              className="border text-black rounded-md px-4 py-2 w-72"
            />
          </div>
          
          {!isRegistering && !isForgetting && (
            <>
                <input
                  type="password"
                  placeholder="Password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  className="border text-black rounded-md px-4 py-2 w-72"
                />
              <div className="mb-2">
                <button onClick={handleLogin} disabled={loading} className="bg-blue-500 text-white px-4 py-2 rounded-md">
                  Login
                </button>
              </div>
              <p onClick={() => setIsRegistering(true)} className="text-blue-500 cursor-pointer mt-2">
                Don't have an account? Register here
              </p>
              <p onClick={() => setIsForgetting(true)} className="text-blue-500 cursor-pointer mt-2">
                Forget Password? Click here
              </p>
            </>
          )}
          {isRegistering && (
            <>
                <input
                  type="password"
                  placeholder="Password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  className="border text-black rounded-md px-4 py-2 w-72"
                />
              <div className="mb-2">
                <button onClick={handleRegister} disabled={loading} className="bg-blue-500 text-white px-4 py-2 rounded-md">
                  Register
                </button>
              </div>
              <input
                type="text"
                placeholder="Enter OTP"
                value={otp}
                onChange={(e) => setOtp(e.target.value)}
                className="border text-black rounded-md px-4 py-2 w-72 mb-2"
              />
              <button onClick={handleVerifyOTP} disabled={loading} className="bg-blue-500 text-white px-4 py-2 rounded-md">
                Verify OTP
              </button>
              <p onClick={() => setIsRegistering(false)} className="text-blue-500 cursor-pointer mt-2">
                Already registered? Login here
              </p>
              <button onClick={handleResendOTP} disabled={remainingTime !== 0} className={`text-blue-500 cursor-pointer mt-2 ${remainingTime !== 0 && 'opacity-50'}`}>
                Resend OTP {remainingTime !== 0 && `(${remainingTime}s)`}
              </button>
              <div className="mb-2">
        </div>
            </>
          )}
          {isForgetting && (
            <>
                <input
                  type="password"
                  placeholder="New Password"
                  value={newPassword}
                  onChange={(e) => setNewPassword(e.target.value)}
                  className="border text-black rounded-md px-4 py-2 w-72"
                />
              <div className="mb-2">
                <button onClick={handleForgetPassword} disabled={loading} className="bg-blue-500 text-white px-4 py-2 rounded-md">
                  Send OTP
                </button>
              </div>
              
              <input
                type="text"
                placeholder="Enter OTP"
                value={otp}
                onChange={(e) => setOtp(e.target.value)}
                className="border text-black rounded-md px-4 py-2 w-72 mb-2"
              />
              <button onClick={handleVerifyOTP} disabled={loading} className="bg-blue-500 text-white px-4 py-2 rounded-md">
                Verify OTP
              </button>
              <button onClick={handleResendOTP} disabled={remainingTime !== 0} className={`text-blue-500 cursor-pointer mt-2 ${remainingTime !== 0 && 'opacity-50'}`}>
                Resend OTP {remainingTime !== 0 && `(${remainingTime}s)`}
              </button>
              <p onClick={() => {
                setIsRegistering(false);
                setIsForgetting(false);
              }} className="text-blue-500 cursor-pointer mt-2">
                Login here
              </p>
            </>
          )}
          {error && <p style={{ color: 'red' }}>{error}</p>}
        </div>
      </div>
    </div>
  );
};

export default LoginPage;