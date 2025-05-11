package middleware

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
